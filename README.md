# MCP GraphQL

This project is a **Model Context Protocol (MCP) server** designed to interact with GraphQL APIs using the official GraphQL client libraries. It allows users to **list available queries and mutations, describe schema entities, and invoke GraphQL operations**.

---

## 🚀 Features

✅ **Invoke GraphQL Operations**: Execute queries and mutations dynamically.  
✅ **List Queries & Mutations**: Retrieve all available queries and mutations in the GraphQL schema.  
✅ **Describe Schema Entities**: Obtain detailed information about GraphQL operations and types.  
✅ **Set Custom Headers**: Configure and manage authentication or request headers for API calls.

---

## 📋 Requirements

- **Go** 1.23.0 or later
- **A running GraphQL API**

---

## ⚙️ Setup

### 1️⃣ Install the Package
```bash
go install github.com/wricardo/mcp-graphql@latest
```

### 2️⃣ Configure Environment Variables
Set the following environment variables to connect to your GraphQL API:
```bash
export ADDRESS="https://your-graphql-endpoint.com"
```

### 3️⃣ Configure MCP Client Settings
Add the following configuration to your MCP settings:
```json
"mcp-graphql": {
  "command": "mcp-graphql",
  "env": {
    "ADDRESS": "https://your-graphql-endpoint.com"
  },
  "disabled": false,
  "autoApprove": []
}
```

---

## ▶️ Usage
Run the MCP server:
```bash
mcp-graphql
```

---

## 🛠️ Tools

### 🔹 **invoke_graphql**
Execute a GraphQL operation (query or mutation).

#### 📌 Parameters:
- `operation` (**required**): The GraphQL query or mutation string.
- `variables` (**optional**): A JSON-encoded string representing query variables.

#### 📌 Example:
```json
{
  "operation": "query { jobs { id name } }"
}
```

---

### 🔹 **list_queries**
Retrieve all available queries in the GraphQL schema.

#### 📌 Parameters:
- None

#### 📌 Example Response:
```json
{
  "queries": [
    "healthcheck(input: String!): String!",
    "candidate(id: String!): Candidate"
  ]
}
```

---

### 🔹 **list_mutations**
Retrieve all available mutations in the GraphQL schema.

#### 📌 Parameters:
- None

#### 📌 Example Response:
```json
{
  "mutations": [
    "createCandidate(input: CandidateInput!): Candidate!",
    "updateInterviewScorecard(id: String!, input: ScorecardInput!): InterviewScorecard!"
  ]
}
```

---

### 🔹 **describe**
Retrieve detailed information about specified GraphQL operations or types.

#### 📌 Parameters:
- `entities` (**required**): A comma-separated list of GraphQL types or operations.

#### 📌 Example:
```json
{
  "entities": "query.jobs,type.JobQueryParams,JobsPage,job"
}
```

---

### 🔹 **set_headers**
Set or overwrite HTTP headers for GraphQL requests.

#### 📌 Parameters:
- `headers` (**required**): JSON-encoded headers.

#### 📌 Example:
```json
{
  "headers": "{\"Authorization\": \"Bearer token123\", \"X-API-Key\": \"abc123\"}"
}
```
