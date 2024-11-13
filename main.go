package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/compliance-framework/plugin-azure-tags/format"

	policyManager "github.com/chris-cmsoft/concom/policy-manager"
	"github.com/chris-cmsoft/concom/runner"
	"github.com/chris-cmsoft/concom/runner/proto"
	"github.com/google/uuid"
	"github.com/hashicorp/go-hclog"
	goplugin "github.com/hashicorp/go-plugin"
)

type AzureCliConfig struct {
	SubscriptionId string `json:"subscriptionid" yaml:"subscriptionid"`
	ClientId       string `json:"clientid" yaml:"clientid"`
	TenantId       string `json:"tenantid" yaml:"tenantid"`
}

type AzureTagsPlugin struct {
	logger     hclog.Logger
	data       map[string]map[string]interface{}
	credential *azidentity.ClientSecretCredential
	cliConfig  AzureCliConfig
}

func (l *AzureTagsPlugin) Configure(req *proto.ConfigureRequest) (*proto.ConfigureResponse, error) {
	azureCliConfig := AzureCliConfig{
		SubscriptionId: req.Config["SubscriptionId"],
		ClientId:       req.Config["ClientId"],
		TenantId:       req.Config["TenantId"],
	}

	// Get environment variable for the secret
	clientSecret := os.Getenv("AZURE_CLIENT_SECRET")

	if azureCliConfig.ClientId == "" || clientSecret == "" || azureCliConfig.TenantId == "" {
		return nil, fmt.Errorf("one or more environment variables are not set")
	}

	cred, err := azidentity.NewClientSecretCredential(azureCliConfig.TenantId, azureCliConfig.ClientId, clientSecret, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain a credential: %v", err)
	}
	l.credential = cred
	l.cliConfig = azureCliConfig
	return &proto.ConfigureResponse{}, nil
}

func (l *AzureTagsPlugin) PrepareForEval(req *proto.PrepareForEvalRequest) (*proto.PrepareForEvalResponse, error) {
	tags, err := format.GetVirtualMachineTags(l.credential, l.cliConfig.SubscriptionId)
	if err != nil {
		return nil, err
	}
	l.data = tags
	return &proto.PrepareForEvalResponse{}, nil
}

func (l *AzureTagsPlugin) Eval(request *proto.EvalRequest) (*proto.EvalResponse, error) {
	ctx := context.TODO()
	start_time := time.Now().Format(time.RFC3339)

	// The Policy Manager aggregates much of the policy execution and output structuring.
	policy_manager := policyManager.New(ctx, l.logger, request.BundlePath)

	response := runner.NewCallableEvalResponse()

	// Iterate over the VMs, and check each against policies
	for machineId, tags := range l.data {
		results, err := policy_manager.Execute(ctx, "azure_tags", tags)

		if err != nil {
			return &proto.EvalResponse{}, err
		}

		for _, result := range results {

			// There are no violations reported from the policies.
			// We'll send the observation back to the agent
			if len(result.Violations) == 0 {
				response.AddObservation(&proto.Observation{
					Id:          uuid.New().String(),
					Title:       "The plugin succeeded. No compliance issues to report.",
					Description: "The plugin policies did not return any violations. The configuration is in compliance with policies.",
					Collected:   time.Now().Format(time.RFC3339),
					Expires:     time.Now().AddDate(0, 1, 0).Format(time.RFC3339), // Add one month for the expiration
					RelevantEvidence: []*proto.Evidence{
						{
							Description: fmt.Sprintf("Policy %v was evaluated, and no violations were found on machineId: %s", result.Policy.Package.PurePackage(), machineId),
						},
					},
				})
			}

			// There are violations in the policy checks.
			// We'll send these observations back to the agent
			if len(result.Violations) > 0 {
				observation := &proto.Observation{
					Id:          uuid.New().String(),
					Title:       fmt.Sprintf("The plugin found violations for policy %s on machineId: %s", result.Policy.Package.PurePackage(), "ARN:12345"),
					Description: fmt.Sprintf("Observed %d violation(s) for policy %s within the Plugin on machineId: %s.", len(result.Violations), result.Policy.Package.PurePackage(), "ARN:12345"),
					Collected:   time.Now().Format(time.RFC3339),
					Expires:     time.Now().AddDate(0, 1, 0).Format(time.RFC3339), // Add one month for the expiration
					RelevantEvidence: []*proto.Evidence{
						{
							Description: fmt.Sprintf("Policy %v was evaluated, and %d violations were found on machineId: %s", result.Policy.Package.PurePackage(), len(result.Violations), machineId),
						},
					},
				}
				response.AddObservation(observation)

				for _, violation := range result.Violations {
					response.AddFinding(&proto.Finding{
						Id:                  uuid.New().String(),
						Title:               violation.GetString("title", fmt.Sprintf("Validation on %s failed with violation %v", result.Policy.Package.PurePackage(), violation)),
						Description:         violation.GetString("description", ""),
						Remarks:             violation.GetString("remarks", ""),
						RelatedObservations: []string{observation.Id},
					})
				}

			}
		}
	}

	response.AddLogEntry(&proto.LogEntry{
		Title: "Plugin checks completed",
		Start: start_time,
		End:   time.Now().Format(time.RFC3339),
	})

	return response.Result(), nil
}

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Debug,
		JSONFormat: true,
	})

	azureTags := &AzureTagsPlugin{
		logger: logger,
	}
	// pluginMap is the map of plugins we can dispense.
	logger.Debug("initiating plugin")

	goplugin.Serve(&goplugin.ServeConfig{
		HandshakeConfig: runner.HandshakeConfig,
		Plugins: map[string]goplugin.Plugin{
			"runner": &runner.RunnerGRPCPlugin{
				Impl: azureTags,
			},
		},
		GRPCServer: goplugin.DefaultGRPCServer,
	})
}
