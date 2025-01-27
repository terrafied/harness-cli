package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

func applyGitopsApplications(c *cli.Context) error {
	filePath := c.String("file")
	baseURL := getBaseUrl(c, GITOPS_BASE_URL)
	if filePath == "" {
		fmt.Println("Please enter valid filename")
		return nil
	}
	fmt.Println("Trying to create or update gitops-application using the yaml=",
		getColoredText(filePath, color.FgCyan))
	var content, _ = readFromFile(c.String("file"))
	agentIdentifier = c.String("agent-identifier")
	if agentIdentifier == "" || agentIdentifier == GITOPS_AGENT_IDENTIFIER_PLACEHOLDER {
		agentIdentifier = TextInput("Enter a valid AgentIdentifier:")
	}
	content = replacePlaceholderValues(content, GITOPS_AGENT_IDENTIFIER_PLACEHOLDER, agentIdentifier)
	baseURL = baseURL + agentIdentifier
	requestBody := getJsonFromYaml(content)
	if requestBody == nil {
		println(getColoredText("Please enter valid repository yaml file", color.FgRed))
	}
	orgIdentifier := valueToString(GetNestedValue(requestBody, "gitops", "orgIdentifier").(string))
	projectIdentifier := valueToString(GetNestedValue(requestBody, "gitops", "projectIdentifier").(string))
	clusterIdentifier := valueToString(GetNestedValue(requestBody, "gitops", "clusterIdentifier").(string))
	repoIdentifier := valueToString(GetNestedValue(requestBody, "gitops", "repoIdentifier").(string))

	createOrUpdateApplicationURL := GetUrlWithQueryParams("", baseURL, GITOPS_APPLICATION_ENDPOINT, map[string]string{
		"accountIdentifier": cliCdRequestData.Account,
		"orgIdentifier":     orgIdentifier,
		"projectIdentifier": projectIdentifier,
		"clusterIdentifier": clusterIdentifier,
		"repoIdentifier":    repoIdentifier,
	})

	applicationName := valueToString(GetNestedValue(requestBody, "gitops", "name").(string))
	syncApplicationURL := GetUrlWithQueryParams("", baseURL,
		fmt.Sprintf(GITOPS_APPLICATION_ENDPOINT+"/%s", applicationName+"/sync"), map[string]string{
			"routingId":         cliCdRequestData.Account,
			"accountIdentifier": cliCdRequestData.Account,
			"orgIdentifier":     orgIdentifier,
			"projectIdentifier": projectIdentifier,
		})

	extraParams := map[string]string{
		"agentIdentifier": agentIdentifier,
	}
	entityExists := getEntity(baseURL, fmt.Sprintf(GITOPS_APPLICATION_ENDPOINT+"/%s", applicationName),
		projectIdentifier, orgIdentifier, extraParams)
	var _ ResponseBody
	var err error

	if !entityExists {
		println("Creating GitOps-Application with id: ", getColoredText(applicationName, color.FgGreen))
		applicationPayload := createGitOpsApplicationPayload(requestBody)
		_, err = Post(createOrUpdateApplicationURL, cliCdRequestData.AuthToken, applicationPayload, CONTENT_TYPE_JSON, nil)
		if err == nil {
			println(getColoredText("Successfully created GitOps-Application with id= ", color.FgGreen) +
				getColoredText(applicationName, color.FgBlue))
			return nil
		}
	} else {
		println("Found GitOps Application with id=", getColoredText(applicationName, color.FgCyan))
		println("Updating details of GitOps Application with id=", getColoredText(applicationName, color.FgBlue))

		var appPUTUrl = GetUrlWithQueryParams("", baseURL,
			fmt.Sprintf(GITOPS_APPLICATION_ENDPOINT+"/%s", applicationName), map[string]string{
				"routingId":         cliCdRequestData.Account,
				"accountIdentifier": cliCdRequestData.Account,
				"orgIdentifier":     orgIdentifier,
				"projectIdentifier": projectIdentifier,
				"repoIdentifier":    repoIdentifier,
				"clusterIdentifier": clusterIdentifier,
			})
		newAppPayload := createGitOpsApplicationPUTPayload(requestBody)
		syncPayload := createGitOpsApplicationPayload(requestBody)
		println("Syncing the GitOps Application before updating the spec:", getColoredText(applicationName, color.FgGreen))
		_, err = Post(syncApplicationURL, cliCdRequestData.AuthToken, syncPayload, CONTENT_TYPE_JSON, nil)
		if err == nil {
			println(getColoredText("Successfully synced GitOps app with id= ", color.FgGreen) +
				getColoredText(applicationName, color.FgBlue))
			_, err = Put(appPUTUrl, cliCdRequestData.AuthToken, newAppPayload, CONTENT_TYPE_JSON, nil)
			if err == nil {
				println(getColoredText("Successfully updated GitOps app with id= ", color.FgGreen) +
					getColoredText(applicationName, color.FgBlue))
				return nil
			}
			return nil
		}
	}

	return nil
}

func createGitOpsApplicationPayload(requestBody map[string]interface{}) GitOpsApplication {
	newApplication := GitOpsApplication{
		Application: Application{
			Metadata: Metadata{
				Labels: Labels{
					Envref:     valueToString(GetNestedValue(requestBody, "gitops", "application", "metadata", "labels", "harness.io/envRef").(string)),
					Serviceref: valueToString(GetNestedValue(requestBody, "gitops", "application", "metadata", "labels", "harness.io/serviceRef").(string)),
				},
				Name:        valueToString(GetNestedValue(requestBody, "gitops", "name").(string)),
				Namespace:   valueToString(GetNestedValue(requestBody, "gitops", "application", "metadata", "namespace").(string)),
				ClusterName: valueToString(GetNestedValue(requestBody, "gitops", "application", "metadata", "clusterName").(string)),
			},
			Spec: Spec{
				Source: Source{
					RepoURL:        valueToString(GetNestedValue(requestBody, "gitops", "application", "spec", "source", "repoURL").(string)),
					Path:           valueToString(GetNestedValue(requestBody, "gitops", "application", "spec", "source", "path").(string)),
					TargetRevision: valueToString(GetNestedValue(requestBody, "gitops", "application", "spec", "source", "targetRevision").(string)),
				},
				Destination: Destination{
					Server:    valueToString(GetNestedValue(requestBody, "gitops", "application", "spec", "destination", "server").(string)),
					Namespace: valueToString(GetNestedValue(requestBody, "gitops", "application", "spec", "destination", "namespace").(string)),
				},
			},
		},
	}
	return newApplication
}

func createGitOpsApplicationPUTPayload(requestBody map[string]interface{}) GitOpsApplication {
	putApp := GitOpsApplication{
		Application{
			Metadata: Metadata{
				Namespace: valueToString(GetNestedValue(requestBody, "gitops", "application", "metadata", "namespace").(string)),
				Labels: Labels{
					Serviceref: valueToString(GetNestedValue(requestBody, "gitops", "application", "metadata", "labels", "harness.io/serviceRef").(string)),
					Envref:     valueToString(GetNestedValue(requestBody, "gitops", "application", "metadata", "labels", "harness.io/envRef").(string)),
				},
			},
			Spec: Spec{
				Source: Source{
					RepoURL:        valueToString(GetNestedValue(requestBody, "gitops", "application", "spec", "source", "repoURL").(string)),
					Path:           valueToString(GetNestedValue(requestBody, "gitops", "application", "spec", "source", "path").(string)),
					TargetRevision: valueToString(GetNestedValue(requestBody, "gitops", "application", "spec", "source", "targetRevision").(string)),
				},
				Destination: Destination{
					Server:    valueToString(GetNestedValue(requestBody, "gitops", "application", "spec", "destination", "server").(string)),
					Namespace: valueToString(GetNestedValue(requestBody, "gitops", "application", "spec", "destination", "namespace").(string)),
				},
			},
		},
	}
	return putApp
}
