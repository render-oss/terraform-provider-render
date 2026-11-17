package resource_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"terraform-provider-render/internal/provider"
	"terraform-provider-render/internal/provider/common"
	postgresresource "terraform-provider-render/internal/provider/postgres/resource"
)

type replicaRemovalAPI struct {
	mu            sync.Mutex
	database      map[string]any
	updates       []json.RawMessage
	creates       int
	deletes       int
	rejectUpdate  bool
	ignoreRemoval bool
}

func (a *replicaRemovalAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	switch {
	case r.URL.Path == "/postgres" && r.Method == http.MethodPost:
		a.creates++
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		replicas, err := replicaResponse(body["readReplicas"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		a.database = map[string]any{
			"id": "dpg-primary", "name": body["name"],
			"databaseName": "appdb", "databaseUser": "appuser",
			"plan": body["plan"], "region": body["region"], "version": body["version"],
			"status": "available", "role": "primary", "diskSizeGB": 20,
			"highAvailabilityEnabled": false, "ipAllowList": []any{},
			"readReplicas": replicas,
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(a.database)
	case r.URL.Path == "/postgres/dpg-primary" && r.Method == http.MethodPatch:
		var body map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		a.updates = append(a.updates, append(json.RawMessage(nil), body["readReplicas"]...))
		if a.rejectUpdate {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("{\"message\":\"replica update rejected\"}"))
			return
		}
		if raw, exists := body["readReplicas"]; exists && string(raw) != "null" {
			var replicas []any
			if err := json.Unmarshal(raw, &replicas); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if !a.ignoreRemoval || len(replicas) > 0 {
				values, err := replicaResponse(replicas)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				a.database["readReplicas"] = values
			}
		}
		if raw, exists := body["name"]; exists {
			var name string
			if err := json.Unmarshal(raw, &name); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			a.database["name"] = name
		}
		_ = json.NewEncoder(w).Encode(a.database)
	case r.URL.Path == "/postgres/dpg-primary" && r.Method == http.MethodGet:
		if a.database == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(a.database)
	case r.URL.Path == "/postgres/dpg-primary" && r.Method == http.MethodDelete:
		a.deletes++
		a.database = nil
		w.WriteHeader(http.StatusNoContent)
	case r.URL.Path == "/postgres/dpg-primary/connection-info" && r.Method == http.MethodGet:
		_, _ = w.Write([]byte("{\"password\":\"fake-password\",\"externalConnectionString\":\"postgres://fake\",\"internalConnectionString\":\"postgres://fake\",\"psqlCommand\":\"psql\"}"))
	case strings.Contains(r.URL.Path, "log-stream") && r.Method == http.MethodGet:
		w.WriteHeader(http.StatusNotFound)
	default:
		http.Error(w, "unexpected fake API request: "+r.Method+" "+r.URL.Path, http.StatusNotFound)
	}
}

func replicaResponse(value any) ([]map[string]any, error) {
	replicas := []map[string]any{}
	items, ok := value.([]any)
	if value != nil && !ok {
		return nil, fmt.Errorf("replicas must be an array")
	}
	for _, item := range items {
		input, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("replica must be an object")
		}
		name, ok := input["name"].(string)
		if !ok || name == "" {
			return nil, fmt.Errorf("replica name is required")
		}
		replicas = append(replicas, map[string]any{
			"name": input["name"], "id": "dpg-" + name,
			"parameterOverrides": input["parameterOverrides"],
		})
	}
	return replicas, nil
}

func replicaConfig(name, replicas string) string {
	return fmt.Sprintf("provider \"render\" {}\nresource \"render_postgres\" \"test\" {\nname = %q\nplan = \"pro_4gb\"\nregion = \"oregon\"\nversion = \"15\"\ndatabase_name = \"appdb\"\ndatabase_user = \"appuser\"\ndisk_size_gb = 20\n%s\n}\n", name, replicas)
}

func replicaProviders(api *httptest.Server) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"render": providerserver.NewProtocol6WithError(provider.New("test",
			provider.WithHost(api.URL), provider.WithHTTPClient(api.Client()),
			provider.WithAPIKey("fake-key"), provider.WithOwnerID("fake-owner"),
			provider.WithPoller(&common.TestPoller),
		)()),
	}
}

func TestAccPostgresReplicaRemoval(t *testing.T) {
	for _, tt := range []struct {
		name     string
		replicas string
	}{
		{name: "empty_set", replicas: "read_replicas = []"},
		{name: "null", replicas: "read_replicas = null"},
		{name: "omitted"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fake := &replicaRemovalAPI{}
			api := httptest.NewServer(fake)
			defer api.Close()
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: replicaProviders(api),
				Steps: []resource.TestStep{
					{Config: replicaConfig("replica-test", "read_replicas = [{ name = \"replica-a\" }, { name = \"replica-b\" }]"),
						Check: resource.TestCheckResourceAttr("render_postgres.test", "read_replicas.#", "2")},
					{Config: replicaConfig("renamed-primary", "read_replicas = [{ name = \"replica-a\" }, { name = \"replica-b\" }]"),
						Check: resource.TestCheckResourceAttr("render_postgres.test", "read_replicas.#", "2")},
					{Config: replicaConfig("renamed-primary", "read_replicas = [{ name = \"replica-a\" }]"),
						Check: resource.TestCheckResourceAttr("render_postgres.test", "read_replicas.#", "1")},
					{Config: replicaConfig("renamed-primary", tt.replicas),
						Check: resource.TestCheckResourceAttr("render_postgres.test", "id", "dpg-primary")},
					{Config: replicaConfig("renamed-primary", tt.replicas), PlanOnly: true},
				},
			})
			fake.mu.Lock()
			defer fake.mu.Unlock()
			if len(fake.updates) < 3 {
				t.Fatalf("expected unchanged, partial and full replica updates, got %d", len(fake.updates))
			}
			if got := strings.TrimSpace(string(fake.updates[len(fake.updates)-1])); got != "[]" {
				t.Errorf("removal request = %s, want []", got)
			}
			if fake.creates != 1 {
				t.Errorf("primary created %d times, want one initial creation", fake.creates)
			}
			if fake.deletes != 1 {
				t.Errorf("primary deleted %d times, want only final test cleanup", fake.deletes)
			}
		})
	}
}

func TestAccPostgresReplicaRemovalErrors(t *testing.T) {
	for _, tt := range []struct {
		name    string
		reject  bool
		ignore  bool
		message string
	}{
		{name: "API rejection", reject: true, message: "replica update rejected"},
		{name: "API retains replica", ignore: true, message: "inconsistent result after apply"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fake := &replicaRemovalAPI{rejectUpdate: tt.reject, ignoreRemoval: tt.ignore}
			api := httptest.NewServer(fake)
			defer api.Close()
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: replicaProviders(api),
				Steps: []resource.TestStep{
					{Config: replicaConfig("replica-test", "read_replicas = [{ name = \"replica-a\" }]")},
					{Config: replicaConfig("replica-test", "read_replicas = []"), ExpectError: regexp.MustCompile(tt.message)},
				},
			})
		})
	}
}

func TestPostgresUpdateUnknownReplicas(t *testing.T) {
	ctx := context.Background()
	r := postgresresource.NewPostgresResource()
	var schema frameworkresource.SchemaResponse
	r.Schema(ctx, frameworkresource.SchemaRequest{}, &schema)
	objectType, ok := schema.Schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatal("resource schema must be an object")
	}
	attributes := make(map[string]tftypes.Value, len(objectType.AttributeTypes))
	for name, typ := range objectType.AttributeTypes {
		attributes[name] = tftypes.NewValue(typ, nil)
	}
	attributes["read_replicas"] = tftypes.NewValue(objectType.AttributeTypes["read_replicas"], tftypes.UnknownValue)
	var response frameworkresource.UpdateResponse
	r.Update(ctx, frameworkresource.UpdateRequest{
		Plan: tfsdk.Plan{Schema: schema.Schema, Raw: tftypes.NewValue(objectType, attributes)},
	}, &response)
	if !response.Diagnostics.HasError() || !strings.Contains(fmt.Sprint(response.Diagnostics), "read_replicas") {
		t.Fatalf("expected unknown replica diagnostic before using the unconfigured API client, got %v", response.Diagnostics)
	}
}

func TestAccPostgresInitiallyWithoutReplicas(t *testing.T) {
	for _, replicas := range []string{"read_replicas = []", "read_replicas = null", ""} {
		t.Run(replicas, func(t *testing.T) {
			fake := &replicaRemovalAPI{}
			api := httptest.NewServer(fake)
			defer api.Close()
			config := replicaConfig("replica-test", replicas)
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: replicaProviders(api),
				Steps: []resource.TestStep{
					{Config: config, Check: resource.TestCheckResourceAttr("render_postgres.test", "id", "dpg-primary")},
					{Config: config, PlanOnly: true},
				},
			})
		})
	}
}
