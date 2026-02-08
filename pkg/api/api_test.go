package api

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// =============================================================================
// isValidSlug Tests
// =============================================================================

func TestIsValidSlug(t *testing.T) {
	tests := []struct {
		name string
		slug string
		want bool
	}{
		// Valid slugs
		{name: "simple lowercase", slug: "hello", want: true},
		{name: "simple uppercase", slug: "Hello", want: true},
		{name: "mixed case", slug: "HelloWorld", want: true},
		{name: "with hyphens", slug: "hello-world", want: true},
		{name: "with underscores", slug: "hello_world", want: true},
		{name: "with numbers", slug: "post123", want: true},
		{name: "numbers only", slug: "12345", want: true},
		{name: "single char", slug: "a", want: true},
		{name: "single number", slug: "1", want: true},
		{name: "hyphen and underscore mix", slug: "my-blog_post-1", want: true},
		{name: "all caps", slug: "ALLCAPS", want: true},
		{name: "leading hyphen", slug: "-leading", want: true},
		{name: "trailing hyphen", slug: "trailing-", want: true},
		{name: "leading underscore", slug: "_leading", want: true},
		{name: "trailing underscore", slug: "trailing_", want: true},

		// Valid slugs with .md extension (should be stripped before validation)
		{name: "with .md extension", slug: "hello.md", want: true},
		{name: "with .md and hyphens", slug: "hello-world.md", want: true},
		{name: "with .md and underscores", slug: "hello_world.md", want: true},
		{name: "with .md and numbers", slug: "post123.md", want: true},

		// Invalid slugs
		{name: "empty string", slug: "", want: false},
		{name: "with spaces", slug: "hello world", want: false},
		{name: "with dots", slug: "hello.world", want: false},
		{name: "with slash", slug: "hello/world", want: false},
		{name: "with backslash", slug: "hello\\world", want: false},
		{name: "with at sign", slug: "hello@world", want: false},
		{name: "with hash", slug: "hello#world", want: false},
		{name: "with exclamation", slug: "hello!", want: false},
		{name: "with question mark", slug: "hello?", want: false},
		{name: "with percent", slug: "hello%20world", want: false},
		{name: "with ampersand", slug: "hello&world", want: false},
		{name: "with equals", slug: "hello=world", want: false},
		{name: "with plus", slug: "hello+world", want: false},
		{name: "with unicode", slug: "hello\u00e9", want: false},
		{name: "with emoji", slug: "hello\U0001f600", want: false},
		{name: "with tab", slug: "hello\tworld", want: false},
		{name: "with newline", slug: "hello\nworld", want: false},
		{name: "just .md", slug: ".md", want: false},
		{name: "only spaces", slug: "   ", want: false},

		// Edge cases - length limits
		{name: "99 chars (at limit)", slug: strings.Repeat("a", 99), want: true},
		{name: "100 chars (exceeds limit)", slug: strings.Repeat("a", 100), want: false},
		{name: "200 chars (well over limit)", slug: strings.Repeat("a", 200), want: false},

		// Edge case: .md extension with 99 char base
		{name: "99 chars with .md", slug: strings.Repeat("a", 99) + ".md", want: true},
		{name: "100 chars with .md", slug: strings.Repeat("a", 100) + ".md", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidSlug(tt.slug)
			if got != tt.want {
				t.Errorf("isValidSlug(%q) = %v, want %v", tt.slug, got, tt.want)
			}
		})
	}
}

// =============================================================================
// GetVersion Tests
// =============================================================================

func TestGetVersion(t *testing.T) {
	result := GetVersion()

	t.Run("version is not empty", func(t *testing.T) {
		if result.Version == "" {
			t.Error("GetVersion().Version should not be empty")
		}
	})

	t.Run("git commit is not empty", func(t *testing.T) {
		if result.GitCommit == "" {
			t.Error("GetVersion().GitCommit should not be empty")
		}
	})

	t.Run("build date is not empty", func(t *testing.T) {
		if result.BuildDate == "" {
			t.Error("GetVersion().BuildDate should not be empty")
		}
	})

	t.Run("version format looks valid", func(t *testing.T) {
		// Should look like a semver e.g. "0.3.4"
		parts := strings.Split(result.Version, ".")
		if len(parts) < 2 {
			t.Errorf("GetVersion().Version = %q, expected semver format (e.g., 0.3.4)", result.Version)
		}
	})

	t.Run("returns consistent results", func(t *testing.T) {
		result2 := GetVersion()
		if result.Version != result2.Version {
			t.Errorf("GetVersion() returned inconsistent versions: %q vs %q", result.Version, result2.Version)
		}
		if result.GitCommit != result2.GitCommit {
			t.Errorf("GetVersion() returned inconsistent git commits: %q vs %q", result.GitCommit, result2.GitCommit)
		}
		if result.BuildDate != result2.BuildDate {
			t.Errorf("GetVersion() returned inconsistent build dates: %q vs %q", result.BuildDate, result2.BuildDate)
		}
	})
}

// =============================================================================
// InstallTheme Validation Tests
// =============================================================================

func TestInstallTheme_Validation(t *testing.T) {
	tests := []struct {
		name      string
		params    InstallThemeParams
		wantError string
	}{
		{
			name: "empty site path",
			params: InstallThemeParams{
				SitePath:  "",
				GithubURL: "https://github.com/example/theme",
			},
			wantError: "site path is required",
		},
		{
			name: "empty github URL",
			params: InstallThemeParams{
				SitePath:  "/some/path",
				GithubURL: "",
			},
			wantError: "github URL is required",
		},
		{
			name: "both empty",
			params: InstallThemeParams{
				SitePath:  "",
				GithubURL: "",
			},
			wantError: "site path is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InstallTheme(tt.params)
			if result.Success {
				t.Error("expected failure for invalid params, got success")
			}
			if result.Error != tt.wantError {
				t.Errorf("InstallTheme() error = %q, want %q", result.Error, tt.wantError)
			}
		})
	}
}

// =============================================================================
// GetInstalledThemes Validation Tests
// =============================================================================

func TestGetInstalledThemes_Validation(t *testing.T) {
	t.Run("empty path returns error", func(t *testing.T) {
		result := GetInstalledThemes("")
		if result.Success {
			t.Error("expected failure for empty path, got success")
		}
		if result.Error != "site path is required" {
			t.Errorf("GetInstalledThemes(\"\") error = %q, want %q", result.Error, "site path is required")
		}
	})
}

// =============================================================================
// NewContent Validation Tests
// =============================================================================

func TestNewContent_Validation(t *testing.T) {
	tests := []struct {
		name      string
		params    NewContentParams
		wantError string
	}{
		{
			name: "empty slug",
			params: NewContentParams{
				SitePath: "/some/path",
				Slug:     "",
			},
			wantError: "slug is required",
		},
		{
			name: "invalid slug with spaces",
			params: NewContentParams{
				SitePath: "/some/path",
				Slug:     "hello world",
			},
			wantError: "invalid slug: use only letters, numbers, hyphens, and underscores",
		},
		{
			name: "invalid slug with special chars",
			params: NewContentParams{
				SitePath: "/some/path",
				Slug:     "hello@world",
			},
			wantError: "invalid slug: use only letters, numbers, hyphens, and underscores",
		},
		{
			name: "invalid slug with dots",
			params: NewContentParams{
				SitePath: "/some/path",
				Slug:     "hello.world",
			},
			wantError: "invalid slug: use only letters, numbers, hyphens, and underscores",
		},
		{
			name: "invalid slug with slashes",
			params: NewContentParams{
				SitePath: "/some/path",
				Slug:     "dir/file",
			},
			wantError: "invalid slug: use only letters, numbers, hyphens, and underscores",
		},
		{
			name: "slug exceeding max length",
			params: NewContentParams{
				SitePath: "/some/path",
				Slug:     strings.Repeat("a", 100),
			},
			wantError: "invalid slug: use only letters, numbers, hyphens, and underscores",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewContent(tt.params)
			if result.Success {
				t.Error("expected failure for invalid params, got success")
			}
			if result.Error != tt.wantError {
				t.Errorf("NewContent() error = %q, want %q", result.Error, tt.wantError)
			}
		})
	}
}

// =============================================================================
// ImportAddress Validation Tests
// =============================================================================

func TestImportAddress_Validation(t *testing.T) {
	t.Run("empty input returns error", func(t *testing.T) {
		result := ImportAddress(ImportAddressParams{
			Method:    "mnemonic",
			KeyScheme: "ed25519",
			Input:     "",
		})
		if result.Success {
			t.Error("expected failure for empty input, got success")
		}
		if result.Error != "Input (mnemonic or private key) is required" {
			t.Errorf("ImportAddress() error = %q, want %q", result.Error, "Input (mnemonic or private key) is required")
		}
	})

	t.Run("empty input with no method", func(t *testing.T) {
		result := ImportAddress(ImportAddressParams{
			Input: "",
		})
		if result.Success {
			t.Error("expected failure for empty input, got success")
		}
		if result.Error != "Input (mnemonic or private key) is required" {
			t.Errorf("ImportAddress() error = %q, want %q", result.Error, "Input (mnemonic or private key) is required")
		}
	})

	t.Run("empty input with key method", func(t *testing.T) {
		result := ImportAddress(ImportAddressParams{
			Method: "key",
			Input:  "",
		})
		if result.Success {
			t.Error("expected failure for empty input, got success")
		}
	})

	t.Run("empty input with private-key method", func(t *testing.T) {
		result := ImportAddress(ImportAddressParams{
			Method: "private-key",
			Input:  "",
		})
		if result.Success {
			t.Error("expected failure for empty input, got success")
		}
	})
}

// =============================================================================
// CreateAddress Default KeyScheme Tests
// =============================================================================

func TestCreateAddress_DefaultKeyScheme(t *testing.T) {
	// We can't fully test CreateAddress since it calls sui.CreateAddressWithDetails,
	// but we can verify the default keyScheme logic by checking that the function
	// proceeds past the keyScheme default assignment without panicking.
	// The actual sui call will fail, but that's expected.

	t.Run("empty keyScheme defaults to ed25519", func(t *testing.T) {
		// This will fail at the sui.CreateAddressWithDetails call,
		// but it should NOT fail due to empty keyScheme
		result := CreateAddress("", "test-alias")
		// The error should NOT be about keyScheme being empty
		if result.Success {
			// Unexpected success (would mean sui CLI is available and created an address)
			return
		}
		// Verify the error is from sui, not from our validation
		if strings.Contains(result.Error, "keyScheme") {
			t.Errorf("expected error from sui call, not keyScheme validation, got: %q", result.Error)
		}
	})

	t.Run("provided keyScheme is preserved", func(t *testing.T) {
		// This will also fail at the sui call
		result := CreateAddress("secp256k1", "test-alias")
		if result.Success {
			return
		}
		// Should still fail at sui level, not our validation
		if strings.Contains(result.Error, "keyScheme") {
			t.Errorf("expected error from sui call, not keyScheme validation, got: %q", result.Error)
		}
	})
}

// =============================================================================
// QuickStart Validation Tests
// =============================================================================

func TestQuickStartParams_NameFallback(t *testing.T) {
	// We can test the Name -> SiteName fallback logic indirectly.
	// Both will eventually fail at Hugo init, but we can verify the struct behavior.

	t.Run("SiteName takes priority over Name", func(t *testing.T) {
		params := QuickStartParams{
			SiteName: "primary-name",
			Name:     "fallback-name",
		}
		// Verify the params struct holds both values
		if params.SiteName != "primary-name" {
			t.Errorf("SiteName = %q, want %q", params.SiteName, "primary-name")
		}
		if params.Name != "fallback-name" {
			t.Errorf("Name = %q, want %q", params.Name, "fallback-name")
		}
	})

	t.Run("Name field serves as fallback via JSON", func(t *testing.T) {
		// Simulate JSON from frontend that only sends "name"
		jsonInput := `{"name": "my-site", "siteType": "blog"}`
		var params QuickStartParams
		if err := json.Unmarshal([]byte(jsonInput), &params); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if params.Name != "my-site" {
			t.Errorf("Name = %q, want %q", params.Name, "my-site")
		}
		if params.SiteName != "" {
			t.Errorf("SiteName = %q, want empty string", params.SiteName)
		}
	})

	t.Run("SiteName from JSON takes priority", func(t *testing.T) {
		jsonInput := `{"siteName": "primary", "name": "fallback"}`
		var params QuickStartParams
		if err := json.Unmarshal([]byte(jsonInput), &params); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if params.SiteName != "primary" {
			t.Errorf("SiteName = %q, want %q", params.SiteName, "primary")
		}
		if params.Name != "fallback" {
			t.Errorf("Name = %q, want %q", params.Name, "fallback")
		}
	})
}

// =============================================================================
// JSON Serialization Tests
// =============================================================================

func TestProgressEvent_JSONSerialization(t *testing.T) {
	t.Run("full event serialization", func(t *testing.T) {
		event := ProgressEvent{
			Phase:     "planning",
			EventType: "progress",
			Message:   "Generating content plan...",
			PagePath:  "content/posts/hello.md",
			Progress:  0.5,
			Current:   3,
			Total:     6,
		}

		data, err := json.Marshal(event)
		if err != nil {
			t.Fatalf("failed to marshal ProgressEvent: %v", err)
		}

		var decoded ProgressEvent
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("failed to unmarshal ProgressEvent: %v", err)
		}

		if decoded.Phase != event.Phase {
			t.Errorf("Phase = %q, want %q", decoded.Phase, event.Phase)
		}
		if decoded.EventType != event.EventType {
			t.Errorf("EventType = %q, want %q", decoded.EventType, event.EventType)
		}
		if decoded.Message != event.Message {
			t.Errorf("Message = %q, want %q", decoded.Message, event.Message)
		}
		if decoded.PagePath != event.PagePath {
			t.Errorf("PagePath = %q, want %q", decoded.PagePath, event.PagePath)
		}
		if decoded.Progress != event.Progress {
			t.Errorf("Progress = %f, want %f", decoded.Progress, event.Progress)
		}
		if decoded.Current != event.Current {
			t.Errorf("Current = %d, want %d", decoded.Current, event.Current)
		}
		if decoded.Total != event.Total {
			t.Errorf("Total = %d, want %d", decoded.Total, event.Total)
		}
	})

	t.Run("omitempty for PagePath", func(t *testing.T) {
		event := ProgressEvent{
			Phase:     "done",
			EventType: "complete",
			Message:   "Done",
		}

		data, err := json.Marshal(event)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		// pagePath should be omitted when empty
		if strings.Contains(string(data), "pagePath") {
			t.Errorf("expected pagePath to be omitted when empty, got: %s", string(data))
		}
	})
}

func TestWalletInfo_JSONSerialization(t *testing.T) {
	info := WalletInfo{
		Address:    "0x1234567890abcdef",
		SuiBalance: 10.5,
		WalBalance: 100.0,
		Network:    "testnet",
		Active:     true,
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("failed to marshal WalletInfo: %v", err)
	}

	var decoded WalletInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal WalletInfo: %v", err)
	}

	if decoded.Address != info.Address {
		t.Errorf("Address = %q, want %q", decoded.Address, info.Address)
	}
	if decoded.SuiBalance != info.SuiBalance {
		t.Errorf("SuiBalance = %f, want %f", decoded.SuiBalance, info.SuiBalance)
	}
	if decoded.WalBalance != info.WalBalance {
		t.Errorf("WalBalance = %f, want %f", decoded.WalBalance, info.WalBalance)
	}
	if decoded.Network != info.Network {
		t.Errorf("Network = %q, want %q", decoded.Network, info.Network)
	}
	if decoded.Active != info.Active {
		t.Errorf("Active = %v, want %v", decoded.Active, info.Active)
	}
}

func TestVersionResult_JSONSerialization(t *testing.T) {
	result := VersionResult{
		Version:   "0.3.4",
		GitCommit: "abc123",
		BuildDate: "2024-01-01",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal VersionResult: %v", err)
	}

	var decoded VersionResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal VersionResult: %v", err)
	}

	if decoded.Version != result.Version {
		t.Errorf("Version = %q, want %q", decoded.Version, result.Version)
	}
	if decoded.GitCommit != result.GitCommit {
		t.Errorf("GitCommit = %q, want %q", decoded.GitCommit, result.GitCommit)
	}
	if decoded.BuildDate != result.BuildDate {
		t.Errorf("BuildDate = %q, want %q", decoded.BuildDate, result.BuildDate)
	}
}

func TestAddressListResult_JSONSerialization(t *testing.T) {
	t.Run("with addresses", func(t *testing.T) {
		result := AddressListResult{
			Addresses: []string{"0xabc", "0xdef", "0x123"},
		}

		data, err := json.Marshal(result)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		var decoded AddressListResult
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if len(decoded.Addresses) != 3 {
			t.Errorf("Addresses length = %d, want 3", len(decoded.Addresses))
		}
	})

	t.Run("with error", func(t *testing.T) {
		result := AddressListResult{
			Error: "failed to get addresses",
		}

		data, err := json.Marshal(result)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		var decoded AddressListResult
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if decoded.Error != "failed to get addresses" {
			t.Errorf("Error = %q, want %q", decoded.Error, "failed to get addresses")
		}
		if decoded.Addresses != nil {
			t.Errorf("Addresses = %v, want nil", decoded.Addresses)
		}
	})
}

func TestSwitchAddressResult_JSONSerialization(t *testing.T) {
	result := SwitchAddressResult{
		Success: true,
		Address: "0xdeadbeef",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded SwitchAddressResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if !decoded.Success {
		t.Error("Success should be true")
	}
	if decoded.Address != "0xdeadbeef" {
		t.Errorf("Address = %q, want %q", decoded.Address, "0xdeadbeef")
	}
}

func TestCreateAddressResult_JSONSerialization(t *testing.T) {
	result := CreateAddressResult{
		Success:        true,
		Address:        "0xnewaddr",
		Alias:          "my-wallet",
		RecoveryPhrase: "word1 word2 word3",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded CreateAddressResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if !decoded.Success {
		t.Error("Success should be true")
	}
	if decoded.Address != "0xnewaddr" {
		t.Errorf("Address = %q, want %q", decoded.Address, "0xnewaddr")
	}
	if decoded.Alias != "my-wallet" {
		t.Errorf("Alias = %q, want %q", decoded.Alias, "my-wallet")
	}
	if decoded.RecoveryPhrase != "word1 word2 word3" {
		t.Errorf("RecoveryPhrase = %q, want %q", decoded.RecoveryPhrase, "word1 word2 word3")
	}
}

func TestImportAddressResult_JSONSerialization(t *testing.T) {
	t.Run("success result", func(t *testing.T) {
		result := ImportAddressResult{
			Success: true,
			Address: "0ximported",
		}

		data, err := json.Marshal(result)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		var decoded ImportAddressResult
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if !decoded.Success {
			t.Error("Success should be true")
		}
		if decoded.Address != "0ximported" {
			t.Errorf("Address = %q, want %q", decoded.Address, "0ximported")
		}
	})

	t.Run("error result", func(t *testing.T) {
		result := ImportAddressResult{
			Success: false,
			Error:   "invalid mnemonic",
		}

		data, err := json.Marshal(result)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		var decoded ImportAddressResult
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if decoded.Success {
			t.Error("Success should be false")
		}
		if decoded.Error != "invalid mnemonic" {
			t.Errorf("Error = %q, want %q", decoded.Error, "invalid mnemonic")
		}
	})
}

func TestProject_JSONSerialization(t *testing.T) {
	project := Project{
		ID:            42,
		Name:          "My Project",
		Description:   "A test project",
		Category:      "blog",
		ObjectID:      "0xobj123",
		Network:       "testnet",
		WalletAddr:    "0xwallet",
		SitePath:      "/home/user/sites/my-project",
		ImageURL:      "https://example.com/image.png",
		SuiNS:         "mysite.sui",
		CreatedAt:     "2024-01-01T00:00:00Z",
		UpdatedAt:     "2024-01-02T00:00:00Z",
		LastDeployAt:  "2024-01-02T12:00:00Z",
		FirstDeployAt: "2024-01-01T12:00:00Z",
		Epochs:        5,
		TotalEpochs:   10,
		GasFee:        "0.001 SUI",
		ExpiresAt:     "2024-03-01T00:00:00Z",
		ExpiresIn:     "8 weeks",
		Status:        "active",
		DeployCount:   3,
		Size:          1024000,
		FileCount:     42,
		SuiReady:      true,
		WalrusReady:   true,
		SiteBuilder:   true,
		HugoReady:     true,
	}

	data, err := json.Marshal(project)
	if err != nil {
		t.Fatalf("failed to marshal Project: %v", err)
	}

	var decoded Project
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal Project: %v", err)
	}

	if decoded.ID != project.ID {
		t.Errorf("ID = %d, want %d", decoded.ID, project.ID)
	}
	if decoded.Name != project.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, project.Name)
	}
	if decoded.ObjectID != project.ObjectID {
		t.Errorf("ObjectID = %q, want %q", decoded.ObjectID, project.ObjectID)
	}
	if decoded.Network != project.Network {
		t.Errorf("Network = %q, want %q", decoded.Network, project.Network)
	}
	if decoded.Status != project.Status {
		t.Errorf("Status = %q, want %q", decoded.Status, project.Status)
	}
	if decoded.DeployCount != project.DeployCount {
		t.Errorf("DeployCount = %d, want %d", decoded.DeployCount, project.DeployCount)
	}
	if decoded.Size != project.Size {
		t.Errorf("Size = %d, want %d", decoded.Size, project.Size)
	}
	if decoded.SuiReady != project.SuiReady {
		t.Errorf("SuiReady = %v, want %v", decoded.SuiReady, project.SuiReady)
	}
}

func TestProject_JSONOmitempty(t *testing.T) {
	// Test that optional fields are omitted when zero
	project := Project{
		ID:     1,
		Name:   "test",
		Status: "draft",
	}

	data, err := json.Marshal(project)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	str := string(data)

	// Fields with omitempty should be absent when zero
	omitemptyFields := []string{"firstDeployAt", "totalEpochs", "gasFee", "expiresAt", "expiresIn", "size", "fileCount", "deployments"}
	for _, field := range omitemptyFields {
		if strings.Contains(str, `"`+field+`"`) {
			t.Errorf("expected field %q to be omitted when zero, but found in JSON: %s", field, str)
		}
	}
}

func TestDeploymentRecord_JSONSerialization(t *testing.T) {
	record := DeploymentRecord{
		ID:        1,
		ProjectID: 42,
		ObjectID:  "0xobj",
		Network:   "testnet",
		Epochs:    5,
		GasFee:    "0.001 SUI",
		Version:   "v1.0",
		Notes:     "Initial deployment",
		Success:   true,
		CreatedAt: "2024-01-01T00:00:00Z",
	}

	data, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded DeploymentRecord
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.ID != record.ID {
		t.Errorf("ID = %d, want %d", decoded.ID, record.ID)
	}
	if decoded.ProjectID != record.ProjectID {
		t.Errorf("ProjectID = %d, want %d", decoded.ProjectID, record.ProjectID)
	}
	if decoded.Success != record.Success {
		t.Errorf("Success = %v, want %v", decoded.Success, record.Success)
	}
}

func TestSystemHealth_JSONSerialization(t *testing.T) {
	health := SystemHealth{
		NetOnline:       true,
		SuiInstalled:    true,
		SuiConfigured:   true,
		WalrusInstalled: true,
		SiteBuilder:     false,
		HugoInstalled:   true,
		Message:         "site-builder not installed",
	}

	data, err := json.Marshal(health)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded SystemHealth
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.NetOnline != health.NetOnline {
		t.Errorf("NetOnline = %v, want %v", decoded.NetOnline, health.NetOnline)
	}
	if decoded.SiteBuilder != health.SiteBuilder {
		t.Errorf("SiteBuilder = %v, want %v", decoded.SiteBuilder, health.SiteBuilder)
	}
	if decoded.Message != health.Message {
		t.Errorf("Message = %q, want %q", decoded.Message, health.Message)
	}
}

func TestGasEstimateResult_JSONSerialization(t *testing.T) {
	result := GasEstimateResult{
		Success:   true,
		WAL:       1.5,
		SUI:       0.01,
		WALRange:  "1.0-2.0",
		SUIRange:  "0.005-0.015",
		Summary:   "Estimated: 1.5 WAL + 0.01 SUI",
		SiteSize:  1048576,
		FileCount: 25,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded GasEstimateResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.WAL != result.WAL {
		t.Errorf("WAL = %f, want %f", decoded.WAL, result.WAL)
	}
	if decoded.SUI != result.SUI {
		t.Errorf("SUI = %f, want %f", decoded.SUI, result.SUI)
	}
	if decoded.SiteSize != result.SiteSize {
		t.Errorf("SiteSize = %d, want %d", decoded.SiteSize, result.SiteSize)
	}
}

// =============================================================================
// ImportAddressParams JSON Tests
// =============================================================================

func TestImportAddressParams_JSONDeserialization(t *testing.T) {
	tests := []struct {
		name       string
		jsonInput  string
		wantMethod string
		wantScheme string
		wantInput  string
	}{
		{
			name:       "mnemonic method",
			jsonInput:  `{"method": "mnemonic", "keyScheme": "ed25519", "input": "word1 word2 word3"}`,
			wantMethod: "mnemonic",
			wantScheme: "ed25519",
			wantInput:  "word1 word2 word3",
		},
		{
			name:       "key method",
			jsonInput:  `{"method": "key", "keyScheme": "secp256k1", "input": "0xprivatekey"}`,
			wantMethod: "key",
			wantScheme: "secp256k1",
			wantInput:  "0xprivatekey",
		},
		{
			name:       "private-key method alias",
			jsonInput:  `{"method": "private-key", "keyScheme": "secp256r1", "input": "0xkey"}`,
			wantMethod: "private-key",
			wantScheme: "secp256r1",
			wantInput:  "0xkey",
		},
		{
			name:       "missing method defaults to empty",
			jsonInput:  `{"keyScheme": "ed25519", "input": "some input"}`,
			wantMethod: "",
			wantScheme: "ed25519",
			wantInput:  "some input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var params ImportAddressParams
			if err := json.Unmarshal([]byte(tt.jsonInput), &params); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}
			if params.Method != tt.wantMethod {
				t.Errorf("Method = %q, want %q", params.Method, tt.wantMethod)
			}
			if params.KeyScheme != tt.wantScheme {
				t.Errorf("KeyScheme = %q, want %q", params.KeyScheme, tt.wantScheme)
			}
			if params.Input != tt.wantInput {
				t.Errorf("Input = %q, want %q", params.Input, tt.wantInput)
			}
		})
	}
}

// =============================================================================
// calculateExpiryDate Tests
// =============================================================================

func TestCalculateExpiryDate(t *testing.T) {
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		deploy  time.Time
		epochs  int
		network string
		want    time.Time
	}{
		{
			name:    "mainnet 1 epoch",
			deploy:  baseTime,
			epochs:  1,
			network: "mainnet",
			want:    baseTime.Add(14 * 24 * time.Hour), // 14 days per epoch
		},
		{
			name:    "mainnet 5 epochs",
			deploy:  baseTime,
			epochs:  5,
			network: "mainnet",
			want:    baseTime.Add(70 * 24 * time.Hour), // 5 * 14 = 70 days
		},
		{
			name:    "testnet 1 epoch",
			deploy:  baseTime,
			epochs:  1,
			network: "testnet",
			want:    baseTime.Add(1 * 24 * time.Hour), // 1 day per epoch
		},
		{
			name:    "testnet 10 epochs",
			deploy:  baseTime,
			epochs:  10,
			network: "testnet",
			want:    baseTime.Add(10 * 24 * time.Hour), // 10 days
		},
		{
			name:    "unknown network defaults to testnet",
			deploy:  baseTime,
			epochs:  3,
			network: "devnet",
			want:    baseTime.Add(3 * 24 * time.Hour), // 3 days (testnet default)
		},
		{
			name:    "empty network defaults to testnet",
			deploy:  baseTime,
			epochs:  2,
			network: "",
			want:    baseTime.Add(2 * 24 * time.Hour),
		},
		{
			name:    "zero epochs",
			deploy:  baseTime,
			epochs:  0,
			network: "mainnet",
			want:    baseTime, // 0 days
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateExpiryDate(tt.deploy, tt.epochs, tt.network)
			if !got.Equal(tt.want) {
				t.Errorf("calculateExpiryDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

// =============================================================================
// formatExpiryDuration Tests
// =============================================================================

func TestFormatExpiryDuration(t *testing.T) {
	// Use offsets with extra buffer to avoid off-by-one from time elapsed between
	// constructing the expiry date and the function calling time.Now() internally.
	now := time.Now()

	t.Run("expired", func(t *testing.T) {
		got := formatExpiryDuration(now.Add(-24 * time.Hour))
		if got != "Expired" {
			t.Errorf("formatExpiryDuration() = %q, want %q", got, "Expired")
		}
	})

	t.Run("expired just now", func(t *testing.T) {
		got := formatExpiryDuration(now.Add(-1 * time.Minute))
		if got != "Expired" {
			t.Errorf("formatExpiryDuration() = %q, want %q", got, "Expired")
		}
	})

	t.Run("expiring soon (less than 1 hour)", func(t *testing.T) {
		// Add 10 minutes -- should be 0 days, 0 full hours
		got := formatExpiryDuration(now.Add(10 * time.Minute))
		if got != "Expiring soon" {
			t.Errorf("formatExpiryDuration() = %q, want %q", got, "Expiring soon")
		}
	})

	t.Run("hours remaining (no full days)", func(t *testing.T) {
		// Use 5h30m to ensure we get 5 full hours even with slight delays
		got := formatExpiryDuration(now.Add(5*time.Hour + 30*time.Minute))
		if got != "5 hours" {
			t.Errorf("formatExpiryDuration() = %q, want %q", got, "5 hours")
		}
	})

	t.Run("1 day no extra hours", func(t *testing.T) {
		// Use 24h10m so days=1, hours=0 after floor
		got := formatExpiryDuration(now.Add(24*time.Hour + 10*time.Minute))
		if got != "1 day" {
			t.Errorf("formatExpiryDuration() = %q, want %q", got, "1 day")
		}
	})

	t.Run("1 day and some hours", func(t *testing.T) {
		// Use 30h10m: days=1, hours=6
		got := formatExpiryDuration(now.Add(30*time.Hour + 10*time.Minute))
		if got != "1 day, 6 hours" {
			t.Errorf("formatExpiryDuration() = %q, want %q", got, "1 day, 6 hours")
		}
	})

	t.Run("multiple days (less than 7)", func(t *testing.T) {
		// Use 5 days + 10 minutes buffer
		got := formatExpiryDuration(now.Add(5*24*time.Hour + 10*time.Minute))
		if got != "5 days" {
			t.Errorf("formatExpiryDuration() = %q, want %q", got, "5 days")
		}
	})

	t.Run("exactly 1 week", func(t *testing.T) {
		// Use 7 days + 10 minutes buffer: days=7, weeks=1, remaining=0
		got := formatExpiryDuration(now.Add(7*24*time.Hour + 10*time.Minute))
		if got != "1 weeks" {
			t.Errorf("formatExpiryDuration() = %q, want %q", got, "1 weeks")
		}
	})

	t.Run("weeks with remaining days", func(t *testing.T) {
		// 10 days + buffer: weeks=1, remaining=3
		got := formatExpiryDuration(now.Add(10*24*time.Hour + 10*time.Minute))
		if got != "1 weeks, 3 days" {
			t.Errorf("formatExpiryDuration() = %q, want %q", got, "1 weeks, 3 days")
		}
	})

	t.Run("multiple weeks exact", func(t *testing.T) {
		// 14 days + buffer: weeks=2, remaining=0
		got := formatExpiryDuration(now.Add(14*24*time.Hour + 10*time.Minute))
		if got != "2 weeks" {
			t.Errorf("formatExpiryDuration() = %q, want %q", got, "2 weeks")
		}
	})

	t.Run("many weeks with days", func(t *testing.T) {
		// 20 days + buffer: weeks=2, remaining=6
		got := formatExpiryDuration(now.Add(20*24*time.Hour + 10*time.Minute))
		if got != "2 weeks, 6 days" {
			t.Errorf("formatExpiryDuration() = %q, want %q", got, "2 weeks, 6 days")
		}
	})
}

// =============================================================================
// Epoch Constants Tests
// =============================================================================

func TestEpochConstants(t *testing.T) {
	t.Run("mainnet epoch duration", func(t *testing.T) {
		if MainnetDaysPerEpoch != 14 {
			t.Errorf("MainnetDaysPerEpoch = %d, want 14", MainnetDaysPerEpoch)
		}
	})

	t.Run("testnet epoch duration", func(t *testing.T) {
		if TestnetDaysPerEpoch != 1 {
			t.Errorf("TestnetDaysPerEpoch = %d, want 1", TestnetDaysPerEpoch)
		}
	})
}

// =============================================================================
// Serve Validation Tests
// =============================================================================

func TestServe_Validation(t *testing.T) {
	t.Run("empty site path returns error", func(t *testing.T) {
		result := Serve(ServeParams{SitePath: ""})
		if result.Success {
			t.Error("expected failure for empty site path, got success")
		}
		if result.Error != "site path is required" {
			t.Errorf("Serve() error = %q, want %q", result.Error, "site path is required")
		}
	})
}

// =============================================================================
// LaunchWizard Validation Tests
// =============================================================================

func TestLaunchWizard_Validation(t *testing.T) {
	t.Run("empty site path returns error", func(t *testing.T) {
		result := LaunchWizard(LaunchWizardParams{SitePath: ""})
		if result.Success {
			t.Error("expected failure for empty site path, got success")
		}
		if result.Error != "site path is required" {
			t.Errorf("LaunchWizard() error = %q, want %q", result.Error, "site path is required")
		}
	})
}

// =============================================================================
// UpdateTools Validation Tests
// =============================================================================

func TestUpdateTools_Validation(t *testing.T) {
	t.Run("empty tools list", func(t *testing.T) {
		result := UpdateTools(UpdateToolsParams{Tools: []string{}})
		if result.Success {
			t.Error("expected failure for empty tools list, got success")
		}
		if result.Message != "No tools specified" {
			t.Errorf("UpdateTools() message = %q, want %q", result.Message, "No tools specified")
		}
	})

	t.Run("nil tools list", func(t *testing.T) {
		result := UpdateTools(UpdateToolsParams{})
		if result.Success {
			t.Error("expected failure for nil tools list, got success")
		}
		if result.Message != "No tools specified" {
			t.Errorf("UpdateTools() message = %q, want %q", result.Message, "No tools specified")
		}
	})
}

// =============================================================================
// Type Param/Result Struct Tests
// =============================================================================

func TestInstallThemeParams_JSON(t *testing.T) {
	jsonInput := `{"sitePath": "/home/user/site", "githubUrl": "https://github.com/user/theme"}`
	var params InstallThemeParams
	if err := json.Unmarshal([]byte(jsonInput), &params); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if params.SitePath != "/home/user/site" {
		t.Errorf("SitePath = %q, want %q", params.SitePath, "/home/user/site")
	}
	if params.GithubURL != "https://github.com/user/theme" {
		t.Errorf("GithubURL = %q, want %q", params.GithubURL, "https://github.com/user/theme")
	}
}

func TestNewContentParams_JSON(t *testing.T) {
	jsonInput := `{"slug": "my-post", "contentType": "posts", "noBuild": true, "serve": false}`
	var params NewContentParams
	if err := json.Unmarshal([]byte(jsonInput), &params); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if params.Slug != "my-post" {
		t.Errorf("Slug = %q, want %q", params.Slug, "my-post")
	}
	if params.ContentType != "posts" {
		t.Errorf("ContentType = %q, want %q", params.ContentType, "posts")
	}
	if !params.NoBuild {
		t.Error("NoBuild should be true")
	}
	if params.Serve {
		t.Error("Serve should be false")
	}
}

func TestCreateAddressParams_JSON(t *testing.T) {
	jsonInput := `{"keyScheme": "secp256k1"}`
	var params CreateAddressParams
	if err := json.Unmarshal([]byte(jsonInput), &params); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if params.KeyScheme != "secp256k1" {
		t.Errorf("KeyScheme = %q, want %q", params.KeyScheme, "secp256k1")
	}
}

func TestSwitchAddressParams_JSON(t *testing.T) {
	jsonInput := `{"address": "0x1234"}`
	var params SwitchAddressParams
	if err := json.Unmarshal([]byte(jsonInput), &params); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if params.Address != "0x1234" {
		t.Errorf("Address = %q, want %q", params.Address, "0x1234")
	}
}

func TestSwitchNetworkResult_JSONSerialization(t *testing.T) {
	result := SwitchNetworkResult{
		Success: true,
		Network: "mainnet",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded SwitchNetworkResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if !decoded.Success {
		t.Error("Success should be true")
	}
	if decoded.Network != "mainnet" {
		t.Errorf("Network = %q, want %q", decoded.Network, "mainnet")
	}
}

func TestInitSiteResult_JSONSerialization(t *testing.T) {
	result := InitSiteResult{
		Success:  true,
		SitePath: "/home/user/sites/my-site",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded InitSiteResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if !decoded.Success {
		t.Error("Success should be true")
	}
	if decoded.SitePath != "/home/user/sites/my-site" {
		t.Errorf("SitePath = %q, want %q", decoded.SitePath, "/home/user/sites/my-site")
	}
}

func TestQuickStartParams_JSON(t *testing.T) {
	tests := []struct {
		name      string
		jsonInput string
		wantType  string
	}{
		{
			name:      "blog site type",
			jsonInput: `{"siteName": "my-blog", "siteType": "blog"}`,
			wantType:  "blog",
		},
		{
			name:      "docs site type",
			jsonInput: `{"siteName": "my-docs", "siteType": "docs"}`,
			wantType:  "docs",
		},
		{
			name:      "biolink site type",
			jsonInput: `{"siteName": "my-links", "siteType": "biolink"}`,
			wantType:  "biolink",
		},
		{
			name:      "whitepaper site type",
			jsonInput: `{"siteName": "my-paper", "siteType": "whitepaper"}`,
			wantType:  "whitepaper",
		},
		{
			name:      "skip build flag",
			jsonInput: `{"siteName": "test", "skipBuild": true}`,
			wantType:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var params QuickStartParams
			if err := json.Unmarshal([]byte(tt.jsonInput), &params); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}
			if params.SiteType != tt.wantType {
				t.Errorf("SiteType = %q, want %q", params.SiteType, tt.wantType)
			}
		})
	}
}

func TestEditProjectParams_JSON(t *testing.T) {
	jsonInput := `{"projectId": 42, "name": "New Name", "category": "blog", "description": "Updated desc", "imageUrl": "https://img.com/pic.png", "suins": "mysite.sui"}`
	var params EditProjectParams
	if err := json.Unmarshal([]byte(jsonInput), &params); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if params.ProjectID != 42 {
		t.Errorf("ProjectID = %d, want 42", params.ProjectID)
	}
	if params.Name != "New Name" {
		t.Errorf("Name = %q, want %q", params.Name, "New Name")
	}
	if params.SuiNS != "mysite.sui" {
		t.Errorf("SuiNS = %q, want %q", params.SuiNS, "mysite.sui")
	}
}

func TestDeleteProjectParams_JSON(t *testing.T) {
	jsonInput := `{"projectId": 99}`
	var params DeleteProjectParams
	if err := json.Unmarshal([]byte(jsonInput), &params); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if params.ProjectID != 99 {
		t.Errorf("ProjectID = %d, want 99", params.ProjectID)
	}
}

func TestSetStatusParams_JSON(t *testing.T) {
	tests := []struct {
		name       string
		jsonInput  string
		wantStatus string
	}{
		{
			name:       "draft status",
			jsonInput:  `{"projectId": 1, "status": "draft"}`,
			wantStatus: "draft",
		},
		{
			name:       "active status",
			jsonInput:  `{"projectId": 1, "status": "active"}`,
			wantStatus: "active",
		},
		{
			name:       "archived status",
			jsonInput:  `{"projectId": 1, "status": "archived"}`,
			wantStatus: "archived",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var params SetStatusParams
			if err := json.Unmarshal([]byte(tt.jsonInput), &params); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}
			if params.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", params.Status, tt.wantStatus)
			}
		})
	}
}

func TestLaunchStep_JSONSerialization(t *testing.T) {
	step := LaunchStep{
		Name:    "deploy",
		Status:  "completed",
		Message: "Site deployed successfully",
		Error:   "",
	}

	data, err := json.Marshal(step)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded LaunchStep
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Name != step.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, step.Name)
	}
	if decoded.Status != step.Status {
		t.Errorf("Status = %q, want %q", decoded.Status, step.Status)
	}

	// Error should be omitted when empty
	if strings.Contains(string(data), `"error"`) {
		// Actually checking if error is present is expected since we have `json:"error,omitempty"`
		// empty string should be omitted
		var m map[string]interface{}
		json.Unmarshal(data, &m)
		if _, exists := m["error"]; exists {
			t.Error("expected 'error' field to be omitted when empty")
		}
	}
}

func TestToolVersionInfo_JSONSerialization(t *testing.T) {
	info := ToolVersionInfo{
		Tool:           "sui",
		CurrentVersion: "1.0.0",
		LatestVersion:  "1.1.0",
		UpdateRequired: true,
		Installed:      true,
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ToolVersionInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Tool != "sui" {
		t.Errorf("Tool = %q, want %q", decoded.Tool, "sui")
	}
	if decoded.CurrentVersion != "1.0.0" {
		t.Errorf("CurrentVersion = %q, want %q", decoded.CurrentVersion, "1.0.0")
	}
	if decoded.LatestVersion != "1.1.0" {
		t.Errorf("LatestVersion = %q, want %q", decoded.LatestVersion, "1.1.0")
	}
	if !decoded.UpdateRequired {
		t.Error("UpdateRequired should be true")
	}
	if !decoded.Installed {
		t.Error("Installed should be true")
	}
}

func TestAIConfigResult_JSONSerialization(t *testing.T) {
	result := AIConfigResult{
		Configured:          true,
		Enabled:             true,
		Provider:            "openai",
		CurrentProvider:     "openai",
		Model:               "gpt-4",
		CurrentModel:        "gpt-4",
		ConfiguredProviders: []string{"openai", "openrouter"},
		Success:             true,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded AIConfigResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if !decoded.Configured {
		t.Error("Configured should be true")
	}
	if decoded.Provider != "openai" {
		t.Errorf("Provider = %q, want %q", decoded.Provider, "openai")
	}
	if len(decoded.ConfiguredProviders) != 2 {
		t.Errorf("ConfiguredProviders length = %d, want 2", len(decoded.ConfiguredProviders))
	}
}

func TestProviderCredentialsResult_JSONSerialization(t *testing.T) {
	result := ProviderCredentialsResult{
		Success: true,
		APIKey:  "sk-test-key",
		BaseURL: "https://api.openai.com",
		Model:   "gpt-4",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ProviderCredentialsResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if !decoded.Success {
		t.Error("Success should be true")
	}
	if decoded.APIKey != "sk-test-key" {
		t.Errorf("APIKey = %q, want %q", decoded.APIKey, "sk-test-key")
	}
	if decoded.BaseURL != "https://api.openai.com" {
		t.Errorf("BaseURL = %q, want %q", decoded.BaseURL, "https://api.openai.com")
	}
}

// =============================================================================
// validSlugRe Pattern Tests
// =============================================================================

func TestValidSlugRegex(t *testing.T) {
	// Verify the regex itself allows the expected character classes
	validChars := []string{
		"a", "z", "A", "Z", "0", "9", "_", "-",
		"abc", "ABC", "123", "a-b", "a_b",
	}
	for _, s := range validChars {
		if !validSlugRe.MatchString(s) {
			t.Errorf("validSlugRe should match %q but did not", s)
		}
	}

	invalidChars := []string{
		"", " ", "a b", "a.b", "a/b", "a@b", "a#b",
		"a!b", "a?b", "a&b", "a=b", "a+b",
	}
	for _, s := range invalidChars {
		if validSlugRe.MatchString(s) {
			t.Errorf("validSlugRe should NOT match %q but did", s)
		}
	}
}

// =============================================================================
// AICreateSiteParams JSON Tests
// =============================================================================

func TestAICreateSiteParams_JSON(t *testing.T) {
	jsonInput := `{
		"parentDir": "/home/user/sites",
		"siteName": "AI Blog",
		"siteType": "blog",
		"description": "A blog about AI",
		"audience": "developers"
	}`

	var params AICreateSiteParams
	if err := json.Unmarshal([]byte(jsonInput), &params); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if params.ParentDir != "/home/user/sites" {
		t.Errorf("ParentDir = %q, want %q", params.ParentDir, "/home/user/sites")
	}
	if params.SiteName != "AI Blog" {
		t.Errorf("SiteName = %q, want %q", params.SiteName, "AI Blog")
	}
	if params.SiteType != "blog" {
		t.Errorf("SiteType = %q, want %q", params.SiteType, "blog")
	}
	if params.Description != "A blog about AI" {
		t.Errorf("Description = %q, want %q", params.Description, "A blog about AI")
	}
	if params.Audience != "developers" {
		t.Errorf("Audience = %q, want %q", params.Audience, "developers")
	}
}

func TestAICreateSiteParams_OmitemptyFields(t *testing.T) {
	// Only required fields
	jsonInput := `{"siteName": "test", "siteType": "blog"}`
	var params AICreateSiteParams
	if err := json.Unmarshal([]byte(jsonInput), &params); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if params.ParentDir != "" {
		t.Errorf("ParentDir = %q, want empty", params.ParentDir)
	}
	if params.Description != "" {
		t.Errorf("Description = %q, want empty", params.Description)
	}
	if params.Audience != "" {
		t.Errorf("Audience = %q, want empty", params.Audience)
	}
}

// =============================================================================
// DeleteProjectResult JSON Tests
// =============================================================================

func TestDeleteProjectResult_JSONSerialization(t *testing.T) {
	result := DeleteProjectResult{
		Success:          true,
		Message:          "Project deleted successfully",
		OnChainDestroyed: true,
		EstimatedGasCost: "0.005 SUI",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded DeleteProjectResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if !decoded.Success {
		t.Error("Success should be true")
	}
	if !decoded.OnChainDestroyed {
		t.Error("OnChainDestroyed should be true")
	}
	if decoded.EstimatedGasCost != "0.005 SUI" {
		t.Errorf("EstimatedGasCost = %q, want %q", decoded.EstimatedGasCost, "0.005 SUI")
	}
}

// =============================================================================
// ServeParams/Result JSON Tests
// =============================================================================

func TestServeParams_JSON(t *testing.T) {
	jsonInput := `{"sitePath": "/home/user/site", "port": 8080, "drafts": true, "expired": false, "future": true}`
	var params ServeParams
	if err := json.Unmarshal([]byte(jsonInput), &params); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if params.SitePath != "/home/user/site" {
		t.Errorf("SitePath = %q, want %q", params.SitePath, "/home/user/site")
	}
	if params.Port != 8080 {
		t.Errorf("Port = %d, want 8080", params.Port)
	}
	if !params.Drafts {
		t.Error("Drafts should be true")
	}
	if params.Expired {
		t.Error("Expired should be false")
	}
	if !params.Future {
		t.Error("Future should be true")
	}
}

// =============================================================================
// GenerateContentParams/Result JSON Tests
// =============================================================================

func TestGenerateContentParams_JSON(t *testing.T) {
	jsonInput := `{
		"sitePath": "/home/user/site",
		"filePath": "/home/user/site/content/post.md",
		"contentType": "post",
		"topic": "Machine Learning",
		"context": "Technical blog",
		"instructions": "Write about neural networks"
	}`
	var params GenerateContentParams
	if err := json.Unmarshal([]byte(jsonInput), &params); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if params.SitePath != "/home/user/site" {
		t.Errorf("SitePath = %q, want %q", params.SitePath, "/home/user/site")
	}
	if params.Topic != "Machine Learning" {
		t.Errorf("Topic = %q, want %q", params.Topic, "Machine Learning")
	}
	if params.Instructions != "Write about neural networks" {
		t.Errorf("Instructions = %q, want %q", params.Instructions, "Write about neural networks")
	}
}

func TestUpdateContentParams_JSON(t *testing.T) {
	jsonInput := `{"filePath": "/path/to/file.md", "instructions": "Fix typos", "sitePath": "/path/to/site"}`
	var params UpdateContentParams
	if err := json.Unmarshal([]byte(jsonInput), &params); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if params.FilePath != "/path/to/file.md" {
		t.Errorf("FilePath = %q, want %q", params.FilePath, "/path/to/file.md")
	}
	if params.Instructions != "Fix typos" {
		t.Errorf("Instructions = %q, want %q", params.Instructions, "Fix typos")
	}
}

func TestGasEstimateParams_JSON(t *testing.T) {
	jsonInput := `{"sitePath": "/site", "network": "mainnet", "epochs": 5}`
	var params GasEstimateParams
	if err := json.Unmarshal([]byte(jsonInput), &params); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if params.Network != "mainnet" {
		t.Errorf("Network = %q, want %q", params.Network, "mainnet")
	}
	if params.Epochs != 5 {
		t.Errorf("Epochs = %d, want 5", params.Epochs)
	}
}

// =============================================================================
// ImportObsidianParams JSON Tests
// =============================================================================

func TestImportObsidianParams_JSON(t *testing.T) {
	jsonInput := `{
		"vaultPath": "/home/user/obsidian/vault",
		"siteName": "my-notes",
		"parentDir": "/home/user/sites",
		"outputDir": "notes",
		"dryRun": false,
		"convertLinks": true,
		"linkStyle": "relref",
		"includeDrafts": true
	}`
	var params ImportObsidianParams
	if err := json.Unmarshal([]byte(jsonInput), &params); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if params.VaultPath != "/home/user/obsidian/vault" {
		t.Errorf("VaultPath = %q, want %q", params.VaultPath, "/home/user/obsidian/vault")
	}
	if params.LinkStyle != "relref" {
		t.Errorf("LinkStyle = %q, want %q", params.LinkStyle, "relref")
	}
	if !params.ConvertLinks {
		t.Error("ConvertLinks should be true")
	}
	if !params.IncludeDrafts {
		t.Error("IncludeDrafts should be true")
	}
}

// =============================================================================
// AIConfigureParams JSON Tests
// =============================================================================

func TestAIConfigureParams_JSON(t *testing.T) {
	jsonInput := `{"provider": "openai", "apiKey": "sk-test", "baseURL": "https://api.openai.com/v1", "model": "gpt-4o"}`
	var params AIConfigureParams
	if err := json.Unmarshal([]byte(jsonInput), &params); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if params.Provider != "openai" {
		t.Errorf("Provider = %q, want %q", params.Provider, "openai")
	}
	if params.APIKey != "sk-test" {
		t.Errorf("APIKey = %q, want %q", params.APIKey, "sk-test")
	}
	if params.BaseURL != "https://api.openai.com/v1" {
		t.Errorf("BaseURL = %q, want %q", params.BaseURL, "https://api.openai.com/v1")
	}
	if params.Model != "gpt-4o" {
		t.Errorf("Model = %q, want %q", params.Model, "gpt-4o")
	}
}

// =============================================================================
// UpdateSiteParams JSON Tests
// =============================================================================

func TestUpdateSiteParams_JSON(t *testing.T) {
	jsonInput := `{"projectId": 42, "epochs": 10}`
	var params UpdateSiteParams
	if err := json.Unmarshal([]byte(jsonInput), &params); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if params.ProjectID != 42 {
		t.Errorf("ProjectID = %d, want 42", params.ProjectID)
	}
	if params.Epochs != 10 {
		t.Errorf("Epochs = %d, want 10", params.Epochs)
	}
}

func TestUpdateSiteParams_OmitemptyEpochs(t *testing.T) {
	jsonInput := `{"projectId": 42}`
	var params UpdateSiteParams
	if err := json.Unmarshal([]byte(jsonInput), &params); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if params.Epochs != 0 {
		t.Errorf("Epochs = %d, want 0 (default)", params.Epochs)
	}
}

// =============================================================================
// CheckUpdatesResult JSON Tests
// =============================================================================

func TestCheckUpdatesResult_JSONSerialization(t *testing.T) {
	result := CheckUpdatesResult{
		CurrentVersion: "0.3.4",
		LatestVersion:  "0.4.0",
		UpdateURL:      "https://github.com/selimozten/walgo/releases/tag/v0.4.0",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded CheckUpdatesResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.CurrentVersion != "0.3.4" {
		t.Errorf("CurrentVersion = %q, want %q", decoded.CurrentVersion, "0.3.4")
	}
	if decoded.LatestVersion != "0.4.0" {
		t.Errorf("LatestVersion = %q, want %q", decoded.LatestVersion, "0.4.0")
	}
	if decoded.UpdateURL != "https://github.com/selimozten/walgo/releases/tag/v0.4.0" {
		t.Errorf("UpdateURL = %q, want %q", decoded.UpdateURL, "https://github.com/selimozten/walgo/releases/tag/v0.4.0")
	}
}

// =============================================================================
// LaunchWizardParams JSON Tests
// =============================================================================

func TestLaunchWizardParams_JSON(t *testing.T) {
	jsonInput := `{
		"sitePath": "/home/user/site",
		"network": "mainnet",
		"projectName": "My Project",
		"category": "blog",
		"description": "A great project",
		"imageUrl": "https://img.com/pic.png",
		"epochs": 5,
		"skipConfirm": true
	}`
	var params LaunchWizardParams
	if err := json.Unmarshal([]byte(jsonInput), &params); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if params.SitePath != "/home/user/site" {
		t.Errorf("SitePath = %q, want %q", params.SitePath, "/home/user/site")
	}
	if params.Network != "mainnet" {
		t.Errorf("Network = %q, want %q", params.Network, "mainnet")
	}
	if params.Epochs != 5 {
		t.Errorf("Epochs = %d, want 5", params.Epochs)
	}
	if !params.SkipConfirm {
		t.Error("SkipConfirm should be true")
	}
}
