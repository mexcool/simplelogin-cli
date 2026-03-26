package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

const BaseURL = "https://app.simplelogin.io"

// Verbose controls whether HTTP request debug logging is enabled.
// When true, the client logs method, URL, status code, and latency
// to stderr for every request. It is set by the --verbose flag or
// the SL_VERBOSE / SL_DEBUG environment variables.
var Verbose bool

// Client is the SimpleLogin API client.
type Client struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
	verbose    bool
}

// NewClient creates a new API client with the given API key.
// It inherits the current value of the package-level Verbose flag.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: BaseURL,
		verbose: Verbose,
	}
}

// wrapNetworkError inspects common network error types and returns a
// user-friendly message while preserving the original error in the chain.
func wrapNetworkError(err error) error {
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return fmt.Errorf("could not resolve app.simplelogin.io — check your internet connection: %w", err)
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		// Connection refused is a syscall error with the message "connection refused".
		if opErr.Op == "dial" {
			inner := opErr.Unwrap()
			if inner != nil && isConnectionRefused(inner) {
				return fmt.Errorf("could not connect to SimpleLogin API: %w", err)
			}
		}

		// Timeout (either deadline exceeded or the http.Client Timeout fired).
		if opErr.Timeout() {
			return fmt.Errorf("request timed out — check your internet connection: %w", err)
		}

		// TLS errors surface as *tls.AlertError or *tls.RecordHeaderError wrapped
		// inside a net.OpError whose Op is "remote error" or whose inner type is
		// a tls error.
		if opErr.Op == "remote error" {
			return fmt.Errorf("TLS handshake failed — check your network configuration: %w", err)
		}
		var tlsAlert tls.AlertError
		if errors.As(opErr, &tlsAlert) {
			return fmt.Errorf("TLS handshake failed — check your network configuration: %w", err)
		}
		var tlsRecordErr *tls.RecordHeaderError
		if errors.As(opErr, &tlsRecordErr) {
			return fmt.Errorf("TLS handshake failed — check your network configuration: %w", err)
		}
	}

	// Generic timeout check for url.Error (wraps net.OpError on timeouts).
	var urlErr *url.Error
	if errors.As(err, &urlErr) && urlErr.Timeout() {
		return fmt.Errorf("request timed out — check your internet connection: %w", err)
	}

	return err
}

// isConnectionRefused reports whether err represents a "connection refused" syscall error.
func isConnectionRefused(err error) bool {
	// The most portable check is a string match on the syscall error message.
	return err != nil && (err.Error() == "connection refused" ||
		containsString(err.Error(), "connection refused"))
}

// containsString is a simple helper to avoid importing "strings" just for this.
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}

func (c *Client) do(method, path string, body interface{}) ([]byte, int, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	fullURL := c.baseURL + path
	req, err := http.NewRequest(method, fullURL, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authentication", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	elapsed := time.Since(start)
	if err != nil {
		if c.verbose {
			fmt.Fprintf(os.Stderr, "DEBUG: %s %s → error (%dms)\n", method, fullURL, elapsed.Milliseconds())
		}
		return nil, 0, wrapNetworkError(err)
	}
	defer resp.Body.Close()

	if c.verbose {
		fmt.Fprintf(os.Stderr, "DEBUG: %s %s → %d (%dms)\n", method, fullURL, resp.StatusCode, elapsed.Milliseconds())
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response: %w", err)
	}

	return respBody, resp.StatusCode, nil
}

// HandleError processes an API error response.
func HandleError(statusCode int, body []byte, action string) error {
	switch statusCode {
	case 401:
		return fmt.Errorf("authentication failed. Run: sl auth login")
	case 403:
		var errResp map[string]interface{}
		if err := json.Unmarshal(body, &errResp); err == nil {
			if msg, ok := errResp["error"].(string); ok {
				return fmt.Errorf("%s", msg)
			}
		}
		return fmt.Errorf("forbidden: %s", action)
	case 404:
		return fmt.Errorf("not found: %s", action)
	default:
		var errResp map[string]interface{}
		if err := json.Unmarshal(body, &errResp); err == nil {
			if msg, ok := errResp["error"].(string); ok {
				return fmt.Errorf("failed to %s: %s", action, msg)
			}
		}
		return fmt.Errorf("failed to %s (HTTP %d)", action, statusCode)
	}
}

// --- Auth ---

// UserInfo represents user information.
type UserInfo struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	IsPremium       bool   `json:"is_premium"`
	InTrial         bool   `json:"in_trial"`
	ProfilePicURL   string `json:"profile_picture_url"`
	MaxAliasFreePlan int   `json:"max_alias_free_plan"`
}

// GetUserInfo retrieves the current user's information.
func (c *Client) GetUserInfo() (*UserInfo, []byte, error) {
	body, status, err := c.do("GET", "/api/user_info", nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user info: %w", err)
	}
	if status != 200 {
		return nil, body, HandleError(status, body, "get user info")
	}

	var info UserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, body, fmt.Errorf("failed to parse user info: %w", err)
	}
	return &info, body, nil
}

// Logout logs out the current session.
func (c *Client) Logout() error {
	_, status, err := c.do("GET", "/api/logout", nil)
	if err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}
	if status != 200 {
		return fmt.Errorf("logout failed (HTTP %d)", status)
	}
	return nil
}

// --- Stats ---

// Stats represents account statistics.
type Stats struct {
	NbAlias   int `json:"nb_alias"`
	NbBlock   int `json:"nb_block"`
	NbForward int `json:"nb_forward"`
	NbReply   int `json:"nb_reply"`
}

// GetStats retrieves account statistics.
func (c *Client) GetStats() (*Stats, []byte, error) {
	body, status, err := c.do("GET", "/api/stats", nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get stats: %w", err)
	}
	if status != 200 {
		return nil, body, HandleError(status, body, "get stats")
	}

	var stats Stats
	if err := json.Unmarshal(body, &stats); err != nil {
		return nil, body, fmt.Errorf("failed to parse stats: %w", err)
	}
	return &stats, body, nil
}

// --- Aliases ---

// Alias represents a SimpleLogin alias.
type Alias struct {
	ID            int      `json:"id"`
	Email         string   `json:"email"`
	Name          *string  `json:"name"`
	Enabled       bool     `json:"enabled"`
	CreationDate  string   `json:"creation_date"`
	Note          *string  `json:"note"`
	NbBlock       int      `json:"nb_block"`
	NbForward     int      `json:"nb_forward"`
	NbReply       int      `json:"nb_reply"`
	Pinned        bool     `json:"pinned"`
	DisablePGP    bool     `json:"disable_pgp"`
	Mailboxes     []Mailbox `json:"mailboxes"`
	LatestActivity *Activity `json:"latest_activity"`
	Support       bool     `json:"support_pgp"`
}

// AliasListResponse is the response for listing aliases.
type AliasListResponse struct {
	Aliases []Alias `json:"aliases"`
}

// ListAliases retrieves a page of aliases with optional filters.
func (c *Client) ListAliases(pageID int, pinned, disabled, enabled bool, query string) ([]Alias, []byte, error) {
	path := fmt.Sprintf("/api/v2/aliases?page_id=%d", pageID)
	if pinned {
		path += "&pinned"
	}
	if disabled {
		path += "&disabled"
	}
	if enabled {
		path += "&enabled"
	}
	if query != "" {
		path += "&query=" + url.QueryEscape(query)
	}

	body, status, err := c.do("GET", path, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list aliases: %w", err)
	}
	if status != 200 {
		return nil, body, HandleError(status, body, "list aliases")
	}

	var resp AliasListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, body, fmt.Errorf("failed to parse aliases: %w", err)
	}
	return resp.Aliases, body, nil
}

// ListAllAliases retrieves all aliases by paginating through all pages.
func (c *Client) ListAllAliases(pinned, disabled, enabled bool, query string) ([]Alias, error) {
	var all []Alias
	page := 0
	for {
		aliases, _, err := c.ListAliases(page, pinned, disabled, enabled, query)
		if err != nil {
			return nil, err
		}
		if len(aliases) == 0 {
			break
		}
		all = append(all, aliases...)
		page++
	}
	return all, nil
}

// GetAlias retrieves a single alias by ID.
func (c *Client) GetAlias(id int) (*Alias, []byte, error) {
	body, status, err := c.do("GET", fmt.Sprintf("/api/aliases/%d", id), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get alias: %w", err)
	}
	if status != 200 {
		return nil, body, HandleError(status, body, "get alias")
	}

	var alias Alias
	if err := json.Unmarshal(body, &alias); err != nil {
		return nil, body, fmt.Errorf("failed to parse alias: %w", err)
	}
	return &alias, body, nil
}

// DeleteAlias deletes an alias by ID.
func (c *Client) DeleteAlias(id int) error {
	body, status, err := c.do("DELETE", fmt.Sprintf("/api/aliases/%d", id), nil)
	if err != nil {
		return fmt.Errorf("failed to delete alias: %w", err)
	}
	if status != 200 {
		return HandleError(status, body, "delete alias")
	}
	return nil
}

// ToggleAlias toggles an alias between enabled and disabled.
func (c *Client) ToggleAlias(id int) (bool, []byte, error) {
	body, status, err := c.do("POST", fmt.Sprintf("/api/aliases/%d/toggle", id), nil)
	if err != nil {
		return false, nil, fmt.Errorf("failed to toggle alias: %w", err)
	}
	if status != 200 {
		return false, body, HandleError(status, body, "toggle alias")
	}

	var resp struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return false, body, fmt.Errorf("failed to parse toggle response: %w", err)
	}
	return resp.Enabled, body, nil
}

// UpdateAliasRequest is the request body for updating an alias.
type UpdateAliasRequest struct {
	Note       *string `json:"note,omitempty"`
	Name       *string `json:"name,omitempty"`
	MailboxIDs []int   `json:"mailbox_ids,omitempty"`
	Pinned     *bool   `json:"pinned,omitempty"`
	DisablePGP *bool   `json:"disable_pgp,omitempty"`
}

// UpdateAlias updates an alias.
func (c *Client) UpdateAlias(id int, req *UpdateAliasRequest) error {
	body, status, err := c.do("PATCH", fmt.Sprintf("/api/aliases/%d", id), req)
	if err != nil {
		return fmt.Errorf("failed to update alias: %w", err)
	}
	if status != 200 {
		return HandleError(status, body, "update alias")
	}
	return nil
}

// AliasOptions represents the options for creating a custom alias.
type AliasOptions struct {
	CanCreate  bool     `json:"can_create"`
	Prefixes   string   `json:"prefix_suggestion"`
	Suffixes   []SuffixOption `json:"suffixes"`
}

// SuffixOption represents a suffix option.
type SuffixOption struct {
	Suffix       string `json:"suffix"`
	SignedSuffix string `json:"signed_suffix"`
}

// GetAliasOptions retrieves options for creating a custom alias.
func (c *Client) GetAliasOptions() (*AliasOptions, []byte, error) {
	body, status, err := c.do("GET", "/api/v5/alias/options", nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get alias options: %w", err)
	}
	if status != 200 {
		return nil, body, HandleError(status, body, "get alias options")
	}

	var opts AliasOptions
	if err := json.Unmarshal(body, &opts); err != nil {
		return nil, body, fmt.Errorf("failed to parse alias options: %w", err)
	}
	return &opts, body, nil
}

// CreateCustomAliasRequest is the request body for creating a custom alias.
type CreateCustomAliasRequest struct {
	AliasPrefix  string `json:"alias_prefix"`
	SignedSuffix string `json:"signed_suffix"`
	MailboxIDs   []int  `json:"mailbox_ids"`
	Note         string `json:"note,omitempty"`
	Name         string `json:"name,omitempty"`
}

// CreateCustomAlias creates a custom alias.
func (c *Client) CreateCustomAlias(req *CreateCustomAliasRequest) (*Alias, []byte, error) {
	body, status, err := c.do("POST", "/api/v3/alias/custom/new", req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create alias: %w", err)
	}
	if status != 201 {
		return nil, body, HandleError(status, body, "create alias")
	}

	var alias Alias
	if err := json.Unmarshal(body, &alias); err != nil {
		return nil, body, fmt.Errorf("failed to parse alias: %w", err)
	}
	return &alias, body, nil
}

// CreateRandomAliasRequest is the request body for creating a random alias.
type CreateRandomAliasRequest struct {
	Note string `json:"note,omitempty"`
}

// CreateRandomAlias creates a random alias.
func (c *Client) CreateRandomAlias(note string) (*Alias, []byte, error) {
	var reqBody interface{}
	if note != "" {
		reqBody = &CreateRandomAliasRequest{Note: note}
	}
	body, status, err := c.do("POST", "/api/alias/random/new", reqBody)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create random alias: %w", err)
	}
	if status != 201 {
		return nil, body, HandleError(status, body, "create random alias")
	}

	var alias Alias
	if err := json.Unmarshal(body, &alias); err != nil {
		return nil, body, fmt.Errorf("failed to parse alias: %w", err)
	}
	return &alias, body, nil
}

// Activity represents an alias activity entry.
type Activity struct {
	Action    string `json:"action"`
	Timestamp int64  `json:"timestamp"`
	From      string `json:"from"`
	To        string `json:"to"`
	ReversAlias string `json:"reverse_alias"`
}

// ActivityResponse is the response for alias activities.
type ActivityResponse struct {
	Activities []Activity `json:"activities"`
}

// GetAliasActivities retrieves activities for an alias.
func (c *Client) GetAliasActivities(id, pageID int) ([]Activity, []byte, error) {
	body, status, err := c.do("GET", fmt.Sprintf("/api/aliases/%d/activities?page_id=%d", id, pageID), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get alias activities: %w", err)
	}
	if status != 200 {
		return nil, body, HandleError(status, body, "get alias activities")
	}

	var resp ActivityResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, body, fmt.Errorf("failed to parse activities: %w", err)
	}
	return resp.Activities, body, nil
}

// GetAllAliasActivities retrieves all activities for an alias.
func (c *Client) GetAllAliasActivities(id int) ([]Activity, error) {
	var all []Activity
	page := 0
	for {
		activities, _, err := c.GetAliasActivities(id, page)
		if err != nil {
			return nil, err
		}
		if len(activities) == 0 {
			break
		}
		all = append(all, activities...)
		page++
	}
	return all, nil
}

// --- Contacts ---

// Contact represents a contact for an alias.
type Contact struct {
	ID                 int     `json:"id"`
	Contact            string  `json:"contact"`
	CreationDate       string  `json:"creation_date"`
	CreationTimestamp   int64   `json:"creation_timestamp"`
	LastEmailSentDate  *string `json:"last_email_sent_date"`
	ReverseAlias       string  `json:"reverse_alias"`
	ReverseAliasAddress string `json:"reverse_alias_address"`
	BlockForward       bool    `json:"block_forward"`
	Existed            bool    `json:"existed"`
}

// ContactListResponse is the response for listing contacts.
type ContactListResponse struct {
	Contacts []Contact `json:"contacts"`
}

// GetAliasContacts retrieves contacts for an alias.
func (c *Client) GetAliasContacts(aliasID, pageID int) ([]Contact, []byte, error) {
	body, status, err := c.do("GET", fmt.Sprintf("/api/aliases/%d/contacts?page_id=%d", aliasID, pageID), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get contacts: %w", err)
	}
	if status != 200 {
		return nil, body, HandleError(status, body, "get contacts")
	}

	var resp ContactListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, body, fmt.Errorf("failed to parse contacts: %w", err)
	}
	return resp.Contacts, body, nil
}

// GetAllAliasContacts retrieves all contacts for an alias.
func (c *Client) GetAllAliasContacts(aliasID int) ([]Contact, error) {
	var all []Contact
	page := 0
	for {
		contacts, _, err := c.GetAliasContacts(aliasID, page)
		if err != nil {
			return nil, err
		}
		if len(contacts) == 0 {
			break
		}
		all = append(all, contacts...)
		page++
	}
	return all, nil
}

// CreateContactRequest is the request body for creating a contact.
type CreateContactRequest struct {
	Contact string `json:"contact"`
}

// CreateContact creates a contact for an alias.
func (c *Client) CreateContact(aliasID int, contactEmail string) (*Contact, []byte, error) {
	body, status, err := c.do("POST", fmt.Sprintf("/api/aliases/%d/contacts", aliasID), &CreateContactRequest{Contact: contactEmail})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create contact: %w", err)
	}
	if status != 200 && status != 201 {
		return nil, body, HandleError(status, body, "create contact")
	}

	var contact Contact
	if err := json.Unmarshal(body, &contact); err != nil {
		return nil, body, fmt.Errorf("failed to parse contact: %w", err)
	}
	return &contact, body, nil
}

// DeleteContact deletes a contact by ID.
func (c *Client) DeleteContact(id int) error {
	body, status, err := c.do("DELETE", fmt.Sprintf("/api/contacts/%d", id), nil)
	if err != nil {
		return fmt.Errorf("failed to delete contact: %w", err)
	}
	if status != 200 {
		return HandleError(status, body, "delete contact")
	}
	return nil
}

// ToggleContact toggles blocking/unblocking a contact.
func (c *Client) ToggleContact(id int) (bool, []byte, error) {
	body, status, err := c.do("POST", fmt.Sprintf("/api/contacts/%d/toggle", id), nil)
	if err != nil {
		return false, nil, fmt.Errorf("failed to toggle contact: %w", err)
	}
	if status != 200 {
		return false, body, HandleError(status, body, "toggle contact")
	}

	var resp struct {
		BlockForward bool `json:"block_forward"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return false, body, fmt.Errorf("failed to parse toggle response: %w", err)
	}
	return resp.BlockForward, body, nil
}

// --- Mailboxes ---

// Mailbox represents a mailbox.
type Mailbox struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Default  bool   `json:"default"`
	NbAlias  int    `json:"nb_alias"`
	Verified bool   `json:"verified"`
	CreationTimestamp int64 `json:"creation_timestamp"`
}

// MailboxListResponse is the response for listing mailboxes.
type MailboxListResponse struct {
	Mailboxes []Mailbox `json:"mailboxes"`
}

// ListMailboxes retrieves all mailboxes.
func (c *Client) ListMailboxes() ([]Mailbox, []byte, error) {
	body, status, err := c.do("GET", "/api/v2/mailboxes", nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list mailboxes: %w", err)
	}
	if status != 200 {
		return nil, body, HandleError(status, body, "list mailboxes")
	}

	var resp MailboxListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, body, fmt.Errorf("failed to parse mailboxes: %w", err)
	}
	return resp.Mailboxes, body, nil
}

// CreateMailboxRequest is the request body for creating a mailbox.
type CreateMailboxRequest struct {
	Email string `json:"email"`
}

// CreateMailbox creates a new mailbox.
func (c *Client) CreateMailbox(email string) (*Mailbox, []byte, error) {
	body, status, err := c.do("POST", "/api/mailboxes", &CreateMailboxRequest{Email: email})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create mailbox: %w", err)
	}
	if status != 200 && status != 201 {
		return nil, body, HandleError(status, body, "create mailbox")
	}

	var mailbox Mailbox
	if err := json.Unmarshal(body, &mailbox); err != nil {
		return nil, body, fmt.Errorf("failed to parse mailbox: %w", err)
	}
	return &mailbox, body, nil
}

// DeleteMailboxRequest is the request body for deleting a mailbox.
type DeleteMailboxRequest struct {
	TransferAliasesTo *int `json:"transfer_aliases_to,omitempty"`
}

// DeleteMailbox deletes a mailbox.
func (c *Client) DeleteMailbox(id int, transferTo *int) error {
	var reqBody interface{}
	if transferTo != nil {
		reqBody = &DeleteMailboxRequest{TransferAliasesTo: transferTo}
	}
	body, status, err := c.do("DELETE", fmt.Sprintf("/api/mailboxes/%d", id), reqBody)
	if err != nil {
		return fmt.Errorf("failed to delete mailbox: %w", err)
	}
	if status != 200 {
		return HandleError(status, body, "delete mailbox")
	}
	return nil
}

// UpdateMailboxRequest is the request body for updating a mailbox.
type UpdateMailboxRequest struct {
	Default           *bool   `json:"default,omitempty"`
	Email             *string `json:"email,omitempty"`
	CancelEmailChange *bool   `json:"cancel_email_change,omitempty"`
}

// UpdateMailbox updates a mailbox.
func (c *Client) UpdateMailbox(id int, req *UpdateMailboxRequest) error {
	body, status, err := c.do("PUT", fmt.Sprintf("/api/mailboxes/%d", id), req)
	if err != nil {
		return fmt.Errorf("failed to update mailbox: %w", err)
	}
	if status != 200 {
		return HandleError(status, body, "update mailbox")
	}
	return nil
}

// --- Custom Domains ---

// CustomDomain represents a custom domain.
type CustomDomain struct {
	ID                     int       `json:"id"`
	DomainName             string    `json:"domain_name"`
	CreationDate           string    `json:"creation_date"`
	CreationTimestamp       int64     `json:"creation_timestamp"`
	NbAlias                int       `json:"nb_alias"`
	Verified               bool      `json:"is_verified"`
	CatchAll               bool      `json:"catch_all"`
	Name                   *string   `json:"name"`
	RandomPrefixGeneration bool      `json:"random_prefix_generation"`
	Mailboxes              []Mailbox `json:"mailboxes"`
}

// CustomDomainListResponse is the response for listing custom domains.
type CustomDomainListResponse struct {
	CustomDomains []CustomDomain `json:"custom_domains"`
}

// ListCustomDomains retrieves all custom domains.
func (c *Client) ListCustomDomains() ([]CustomDomain, []byte, error) {
	body, status, err := c.do("GET", "/api/custom_domains", nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list custom domains: %w", err)
	}
	if status != 200 {
		return nil, body, HandleError(status, body, "list custom domains")
	}

	var resp CustomDomainListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, body, fmt.Errorf("failed to parse custom domains: %w", err)
	}
	return resp.CustomDomains, body, nil
}

// UpdateDomainRequest is the request body for updating a custom domain.
type UpdateDomainRequest struct {
	CatchAll               *bool   `json:"catch_all,omitempty"`
	RandomPrefixGeneration *bool   `json:"random_prefix_generation,omitempty"`
	Name                   *string `json:"name,omitempty"`
	MailboxIDs             []int   `json:"mailbox_ids,omitempty"`
}

// UpdateCustomDomain updates a custom domain.
func (c *Client) UpdateCustomDomain(id int, req *UpdateDomainRequest) error {
	body, status, err := c.do("PATCH", fmt.Sprintf("/api/custom_domains/%d", id), req)
	if err != nil {
		return fmt.Errorf("failed to update custom domain: %w", err)
	}
	if status != 200 {
		return HandleError(status, body, "update custom domain")
	}
	return nil
}

// DeletedAlias represents a deleted alias from domain trash.
type DeletedAlias struct {
	Alias        string `json:"alias"`
	DeletionDate string `json:"deletion_date"`
}

// DomainTrashResponse is the response for domain trash.
type DomainTrashResponse struct {
	Aliases []DeletedAlias `json:"aliases"`
}

// GetDomainTrash retrieves deleted aliases for a domain.
func (c *Client) GetDomainTrash(domainID int) ([]DeletedAlias, []byte, error) {
	body, status, err := c.do("GET", fmt.Sprintf("/api/custom_domains/%d/trash", domainID), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get domain trash: %w", err)
	}
	if status != 200 {
		return nil, body, HandleError(status, body, "get domain trash")
	}

	var resp DomainTrashResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, body, fmt.Errorf("failed to parse domain trash: %w", err)
	}
	return resp.Aliases, body, nil
}

// --- Settings ---

// Settings represents user settings.
type Settings struct {
	AliasGenerator          string `json:"alias_generator"`
	Notification            bool   `json:"notification"`
	RandomAliasDefaultDomain string `json:"random_alias_default_domain"`
	SenderFormat            string `json:"sender_format"`
	RandomAliasSuffix       string `json:"random_alias_suffix"`
}

// GetSettings retrieves user settings.
func (c *Client) GetSettings() (*Settings, []byte, error) {
	body, status, err := c.do("GET", "/api/setting", nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get settings: %w", err)
	}
	if status != 200 {
		return nil, body, HandleError(status, body, "get settings")
	}

	var settings Settings
	if err := json.Unmarshal(body, &settings); err != nil {
		return nil, body, fmt.Errorf("failed to parse settings: %w", err)
	}
	return &settings, body, nil
}

// UpdateSettingsRequest is the request body for updating settings.
type UpdateSettingsRequest struct {
	AliasGenerator          *string `json:"alias_generator,omitempty"`
	Notification            *bool   `json:"notification,omitempty"`
	RandomAliasDefaultDomain *string `json:"random_alias_default_domain,omitempty"`
	SenderFormat            *string `json:"sender_format,omitempty"`
	RandomAliasSuffix       *string `json:"random_alias_suffix,omitempty"`
}

// UpdateSettings updates user settings.
func (c *Client) UpdateSettings(req *UpdateSettingsRequest) error {
	body, status, err := c.do("PATCH", "/api/setting", req)
	if err != nil {
		return fmt.Errorf("failed to update settings: %w", err)
	}
	if status != 200 {
		return HandleError(status, body, "update settings")
	}
	return nil
}

// Domain represents an available domain for alias creation.
type Domain struct {
	Domain    string `json:"domain"`
	IsCustom  bool   `json:"is_custom"`
}

// DomainListResponse is the response for listing available domains.
type DomainListResponse []Domain

// GetAvailableDomains retrieves available domains for alias creation.
func (c *Client) GetAvailableDomains() ([]Domain, []byte, error) {
	body, status, err := c.do("GET", "/api/v2/setting/domains", nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get domains: %w", err)
	}
	if status != 200 {
		return nil, body, HandleError(status, body, "get available domains")
	}

	var domains []Domain
	if err := json.Unmarshal(body, &domains); err != nil {
		return nil, body, fmt.Errorf("failed to parse domains: %w", err)
	}
	return domains, body, nil
}

// --- Export ---

// ExportData exports all user data.
func (c *Client) ExportData() ([]byte, error) {
	body, status, err := c.do("GET", "/api/export/data", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to export data: %w", err)
	}
	if status != 200 {
		return nil, HandleError(status, body, "export data")
	}
	return body, nil
}

// ExportAliases exports aliases as CSV.
func (c *Client) ExportAliases() ([]byte, error) {
	body, status, err := c.do("GET", "/api/export/aliases", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to export aliases: %w", err)
	}
	if status != 200 {
		return nil, HandleError(status, body, "export aliases")
	}
	return body, nil
}

// --- Helpers ---

// ResolveAliasID resolves an alias email to its ID by searching.
func (c *Client) ResolveAliasID(idOrEmail string) (int, error) {
	// Try parsing as integer first
	if id, err := strconv.Atoi(idOrEmail); err == nil {
		return id, nil
	}

	// Search by email
	aliases, _, err := c.ListAliases(0, false, false, false, idOrEmail)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve alias: %w", err)
	}

	for _, a := range aliases {
		if a.Email == idOrEmail {
			return a.ID, nil
		}
	}

	return 0, fmt.Errorf("alias not found: %s", idOrEmail)
}
