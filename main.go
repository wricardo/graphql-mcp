package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	// Existing library used for introspection
	"github.com/wricardo/graphql"

	// Machine Box library aliased to "graphqlMB"
	graphqlMB "github.com/machinebox/graphql"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Tool descriptions with best practices, argument specifications, and examples.
const (
	invokeToolDescription = `Execute a GraphQL operation (query or mutation) against your API.

Best Practices:
- Use when you have identified the desired operation (query or mutation) and know what variables (if any) need to be supplied.
- Supply 'operation' as the raw GraphQL operation string.
- Optionally provide 'variables' as a JSON-encoded string if the operation uses variables.

Arguments:
- operation (string, Required): The entire GraphQL query or mutation text.
- variables (string, Optional): A JSON-encoded string representing variables for the operation.

Example Usage:
Request:
  invoke_graphql(
    operation: "mutation CreateCandidate($input: CandidateInput!){ createCandidate(input: $input) { id name } }",
    variables: "{\"input\": {\"name\": \"John Doe\"}}"
  )

Response:
  {
    "data": {
      "createCandidate": {
        "id": "123",
        "name": "John Doe"
      }
    }
  }
`
	// Tool: list_queries
	listQueriesToolDescription = `Retrieve a complete list of all available queries in your GraphQL schema. 
This tool is useful for understanding the structure of your API and identifying available queries before implementation or debugging.

Best Practices:
- Use this tool as the first step to understand your GraphQL schema's query capabilities.
- Employ it to quickly identify available queries before implementing or debugging API calls.
- Helps in validating schema changes and documenting GraphQL APIs.

Arguments:
- None

Example Usage:
Request:
  list_queries()

Response:
  Queries:
  healthcheck(input: String!): String!
  candidate(id: String!): Candidate
  interviewScorecard(id: String!): InterviewScorecard
`

	// Tool: list_mutations
	listMutationsToolDescription = `Retrieve a complete list of all available mutations in your GraphQL schema. 
This tool simplifies the process of locating and understanding mutation operations.

Best Practices:
- Start with this tool to get a high-level view of your schema’s mutation capabilities.
- Use it for quick verification of available mutations after schema updates or during debugging.
- Helps in integration testing by listing all possible state-changing operations.

Arguments:
- None

Example Usage:
Request:
  list_mutations()

Response:
  Mutations:
  createCandidate(input: CandidateInput!): Candidate!
  updateInterviewScorecard(id: String!, input: ScorecardInput!): InterviewScorecard!
`

	// Tool: describe
	describeToolDescription = `Provide detailed insights into many operations or types, 
including structure and functionality.

Best Practices:
- Use this tool to understand the structure and functionality of one or many operations or types.

Arguments:
- entities (string) - A comma-separated list of GraphQL operations or types to describe. (Required)

Example Usage:
Request:
  describe("query.jobs,type.JobQueryParams,JobsPage,job")

Response:
  # jobs (Query)
  Arguments:
    page: Int
    size: Int
    search: String
    params: JobQueryParams
  Return Type: JobsPage

  # JobQueryParams (INPUT_OBJECT)
  Input Fields:
    excludedTalentId: String
    locationType: LocationType
    status: JobStatus

  # JobsPage (OBJECT)
  Fields:
    jobs: []
    pagination: Pagination
`
)

// Replace with your actual GraphQL endpoint
var graphqlEndpoint = os.Getenv("ADDRESS")

func getHeaders() http.Header {
	headersJSON := os.Getenv("GRAPHQL_HEADERS")
	var headers = make(http.Header)
	if headersJSON != "" {
		var tmp map[string]string
		if err := json.Unmarshal([]byte(headersJSON), &tmp); err != nil {
			log.Fatal("Failed to parse headers JSON:", err)
		}
		for k, v := range tmp {
			headers.Set(k, v)
		}
	}
	return headers
}

func main() {
	// Validate environment variables
	if graphqlEndpoint == "" {
		log.Fatal("Environment variable ADDRESS is required")
	}

	// introspect the GraphQL endpoint
	_, err := graphql.Introspect(graphqlEndpoint, getHeaders())
	if err != nil {
		log.Fatal("Failed to introspect GraphQL endpoint:", err)
	}

	// Create a new MCP server
	srv := server.NewMCPServer(
		"graphqlServer", "1.0.0", server.WithLogging(),
	)

	// Register tools
	registerTools(srv)

	// Serve the MCP server over standard I/O
	if err := server.ServeStdio(srv); err != nil {
		log.Fatal("Error serving MCP server:", err)
		os.Exit(1)
	}
}

// registerTools registers the available tools with the MCP server.
// It defines:
//   - list_queries
//   - list_mutations
//   - describe
//   - invoke_graphql
func registerTools(srv *server.MCPServer) {
	// Tool 1: list_queries
	listQueriesTool := mcp.NewTool(
		"list_queries",
		mcp.WithDescription(listQueriesToolDescription),
	)
	srv.AddTool(listQueriesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		queries, err := listGraphQLQueries()
		if err != nil {
			return toolError("Failed to list queries: " + err.Error()), nil
		}
		return toolSuccess(queries), nil
	})

	// Tool 2: list_mutations
	listMutationsTool := mcp.NewTool(
		"list_mutations",
		mcp.WithDescription(listMutationsToolDescription),
	)
	srv.AddTool(listMutationsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		mutations, err := listGraphQLMutations()
		if err != nil {
			return toolError("Failed to list mutations: " + err.Error()), nil
		}
		return toolSuccess(mutations), nil
	})

	// Tool 3: describe
	describeTool := mcp.NewTool(
		"describe",
		mcp.WithDescription(describeToolDescription),
		mcp.WithString("entities", mcp.Description("Comma-separated list of operations or types to describe"), mcp.Required()),
	)
	srv.AddTool(describeTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		entities := request.Params.Arguments["entities"].(string)
		description, err := describeGraphQLEntities(entities)
		if err != nil {
			return toolError("Failed to describe entities: " + err.Error()), nil
		}
		return toolSuccess(description), nil
	})

	// Tool 4: invoke_graphql
	// Uses the Machine Box graphql client (aliased as graphqlMB)
	invokeGraphqlTool := mcp.NewTool(
		"invoke_graphql",
		mcp.WithDescription(invokeToolDescription),
		mcp.WithString("operation", mcp.Description("The entire GraphQL operation (query or mutation)"), mcp.Required()),
		mcp.WithString("variables", mcp.Description("JSON-encoded variables for the operation")),
	)
	srv.AddTool(invokeGraphqlTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		operation := request.Params.Arguments["operation"].(string)
		variablesJSON, _ := request.Params.Arguments["variables"].(string)

		resp, err := invokeGraphQLOperation(ctx, operation, variablesJSON)
		if err != nil {
			return toolError(fmt.Sprintf("Failed to invoke GraphQL operation. Operation: %s variables: %v error: %v", operation, variablesJSON, err)), nil
		}
		return toolSuccess(resp), nil
	})
}

// listGraphQLQueries uses the older "github.com/wricardo/graphql" client for introspection
func listGraphQLQueries() (string, error) {
	res, err := graphql.Introspect(graphqlEndpoint, getHeaders())
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	sb.WriteString("Queries:\n")
	for _, typ := range res.Data.Schema.Queries {
		fieldStr := graphql.PrettyPrintField(typ)
		sb.WriteString(fieldStr + "\n")
	}
	return sb.String(), nil
}

// listGraphQLMutations uses the older "github.com/wricardo/graphql" client for introspection
func listGraphQLMutations() (string, error) {
	res, err := graphql.Introspect(graphqlEndpoint, getHeaders())
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	sb.WriteString("Mutations:\n")
	for _, typ := range res.Data.Schema.Mutations {
		fieldStr := graphql.PrettyPrintField(typ)
		sb.WriteString(fieldStr + "\n")
	}
	return sb.String(), nil
}

// describeGraphQLEntities uses the older "github.com/wricardo/graphql" client for deeper introspection
func describeGraphQLEntities(entities string) (string, error) {
	res, err := graphql.Introspect(graphqlEndpoint, getHeaders())
	if err != nil {
		return "", err
	}
	mapp := graphql.GetSchemaMapString(res.Data.Schema)

	entitiesList := strings.Split(entities, ",")
	var descriptions []string
	for _, entity := range entitiesList {
		entity = strings.TrimSpace(entity)
		if desc, ok := mapp[entity]; ok {
			descriptions = append(descriptions, desc)
		} else {
			var example []string
			for k := range mapp {
				example = append(example, k)
				if len(example) >= 3 {
					break
				}
			}
			return "", fmt.Errorf("entity '%s' not found in schema. Example entities in the schema: %s", entity, strings.Join(example, ", "))
		}
	}
	return strings.Join(descriptions, "\n\n"), nil
}

// invokeGraphQLOperation uses the "github.com/machinebox/graphql" client (aliased as graphqlMB)
// to execute a query or mutation, returning a JSON string of the result.
func invokeGraphQLOperation(ctx context.Context, operation, variablesJSON string) (string, error) {
	// Create a Machine Box GraphQL client
	client := graphqlMB.NewClient(graphqlEndpoint)

	// Build the GraphQL request with the raw operation
	req := graphqlMB.NewRequest(operation)

	// If variables were provided, attach them to the request
	if variablesJSON != "" {
		var vars map[string]interface{}
		if err := json.Unmarshal([]byte(variablesJSON), &vars); err != nil {
			return "", fmt.Errorf("failed to parse variables JSON: %w", err)
		}
		for k, v := range vars {
			req.Var(k, v)
		}
	}

	// Read and decode GraphQL headers from environment variable
	headersJSON := os.Getenv("GRAPHQL_HEADERS")
	var headers map[string]string
	if headersJSON != "" {
		if err := json.Unmarshal([]byte(headersJSON), &headers); err != nil {
			return "", fmt.Errorf("failed to parse headers JSON: %w", err)
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	var result interface{}
	if err := client.Run(ctx, req, &result); err != nil {
		return "", err
	}

	// Marshal the result into a pretty JSON string
	resBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(resBytes), nil
}

// toolSuccess formats a successful tool response
func toolSuccess(message string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []interface{}{mcp.NewTextContent(message)},
		IsError: false,
	}
}

// toolError formats an error tool response
func toolError(message string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []interface{}{mcp.NewTextContent(message)},
		IsError: true,
	}
}
