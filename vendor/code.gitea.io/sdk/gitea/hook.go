// Copyright 2014 The Gogs Authors. All rights reserved.
// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gitea

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	// ErrInvalidReceiveHook FIXME
	ErrInvalidReceiveHook = errors.New("Invalid JSON payload received over webhook")
)

// Hook a hook is a web hook when one repository changed
type Hook struct {
	ID     int64             `json:"id"`
	Type   string            `json:"type"`
	URL    string            `json:"-"`
	Config map[string]string `json:"config"`
	Events []string          `json:"events"`
	Active bool              `json:"active"`
	// swagger:strfmt date-time
	Updated time.Time `json:"updated_at"`
	// swagger:strfmt date-time
	Created time.Time `json:"created_at"`
}

// HookList represents a list of API hook.
type HookList []*Hook

// ListOrgHooks list all the hooks of one organization
func (c *Client) ListOrgHooks(org string) (HookList, error) {
	hooks := make([]*Hook, 0, 10)
	return hooks, c.getParsedResponse("GET", fmt.Sprintf("/orgs/%s/hooks", org), nil, nil, &hooks)
}

// ListRepoHooks list all the hooks of one repository
func (c *Client) ListRepoHooks(user, repo string) (HookList, error) {
	hooks := make([]*Hook, 0, 10)
	return hooks, c.getParsedResponse("GET", fmt.Sprintf("/repos/%s/%s/hooks", user, repo), nil, nil, &hooks)
}

// GetOrgHook get a hook of an organization
func (c *Client) GetOrgHook(org string, id int64) (*Hook, error) {
	h := new(Hook)
	return h, c.getParsedResponse("GET", fmt.Sprintf("/orgs/%s/hooks/%d", org, id), nil, nil, h)
}

// GetRepoHook get a hook of a repository
func (c *Client) GetRepoHook(user, repo string, id int64) (*Hook, error) {
	h := new(Hook)
	return h, c.getParsedResponse("GET", fmt.Sprintf("/repos/%s/%s/hooks/%d", user, repo, id), nil, nil, h)
}

// CreateHookOption options when create a hook
type CreateHookOption struct {
	// required: true
	// enum: gitea,gogs,slack,discord
	Type string `json:"type" binding:"Required"`
	// required: true
	Config map[string]string `json:"config" binding:"Required"`
	Events []string          `json:"events"`
	// default: false
	Active bool `json:"active"`
}

// CreateOrgHook create one hook for an organization, with options
func (c *Client) CreateOrgHook(org string, opt CreateHookOption) (*Hook, error) {
	body, err := json.Marshal(&opt)
	if err != nil {
		return nil, err
	}
	h := new(Hook)
	return h, c.getParsedResponse("POST", fmt.Sprintf("/orgs/%s/hooks", org), jsonHeader, bytes.NewReader(body), h)
}

// CreateRepoHook create one hook for a repository, with options
func (c *Client) CreateRepoHook(user, repo string, opt CreateHookOption) (*Hook, error) {
	body, err := json.Marshal(&opt)
	if err != nil {
		return nil, err
	}
	h := new(Hook)
	return h, c.getParsedResponse("POST", fmt.Sprintf("/repos/%s/%s/hooks", user, repo), jsonHeader, bytes.NewReader(body), h)
}

// EditHookOption options when modify one hook
type EditHookOption struct {
	Config map[string]string `json:"config"`
	Events []string          `json:"events"`
	Active *bool             `json:"active"`
}

// EditOrgHook modify one hook of an organization, with hook id and options
func (c *Client) EditOrgHook(org string, id int64, opt EditHookOption) error {
	body, err := json.Marshal(&opt)
	if err != nil {
		return err
	}
	_, err = c.getResponse("PATCH", fmt.Sprintf("/orgs/%s/hooks/%d", org, id), jsonHeader, bytes.NewReader(body))
	return err
}

// EditRepoHook modify one hook of a repository, with hook id and options
func (c *Client) EditRepoHook(user, repo string, id int64, opt EditHookOption) error {
	body, err := json.Marshal(&opt)
	if err != nil {
		return err
	}
	_, err = c.getResponse("PATCH", fmt.Sprintf("/repos/%s/%s/hooks/%d", user, repo, id), jsonHeader, bytes.NewReader(body))
	return err
}

// DeleteOrgHook delete one hook from an organization, with hook id
func (c *Client) DeleteOrgHook(org string, id int64) error {
	_, err := c.getResponse("DELETE", fmt.Sprintf("/org/%s/hooks/%d", org, id), nil, nil)
	return err
}

// DeleteRepoHook delete one hook from a repository, with hook id
func (c *Client) DeleteRepoHook(user, repo string, id int64) error {
	_, err := c.getResponse("DELETE", fmt.Sprintf("/repos/%s/%s/hooks/%d", user, repo, id), nil, nil)
	return err
}

// Payloader payload is some part of one hook
type Payloader interface {
	SetSecret(string)
	JSONPayload() ([]byte, error)
}

// PayloadUser represents the author or committer of a commit
type PayloadUser struct {
	// Full name of the commit author
	Name string `json:"name"`
	// swagger:strfmt email
	Email    string `json:"email"`
	UserName string `json:"username"`
}

// FIXME: consider using same format as API when commits API are added.
//        applies to PayloadCommit and PayloadCommitVerification

// PayloadCommit represents a commit
type PayloadCommit struct {
	// sha1 hash of the commit
	ID           string                     `json:"id"`
	Message      string                     `json:"message"`
	URL          string                     `json:"url"`
	Author       *PayloadUser               `json:"author"`
	Committer    *PayloadUser               `json:"committer"`
	Verification *PayloadCommitVerification `json:"verification"`
	// swagger:strfmt date-time
	Timestamp time.Time `json:"timestamp"`
}

// PayloadCommitVerification represents the GPG verification of a commit
type PayloadCommitVerification struct {
	Verified  bool   `json:"verified"`
	Reason    string `json:"reason"`
	Signature string `json:"signature"`
	Payload   string `json:"payload"`
}

var (
	_ Payloader = &CreatePayload{}
	_ Payloader = &DeletePayload{}
	_ Payloader = &ForkPayload{}
	_ Payloader = &PushPayload{}
	_ Payloader = &IssuePayload{}
	_ Payloader = &IssueCommentPayload{}
	_ Payloader = &PullRequestPayload{}
	_ Payloader = &RepositoryPayload{}
	_ Payloader = &ReleasePayload{}
)

// _________                        __
// \_   ___ \_______   ____ _____ _/  |_  ____
// /    \  \/\_  __ \_/ __ \\__  \\   __\/ __ \
// \     \____|  | \/\  ___/ / __ \|  | \  ___/
//  \______  /|__|    \___  >____  /__|  \___  >
//         \/             \/     \/          \/

// CreatePayload FIXME
type CreatePayload struct {
	Secret  string      `json:"secret"`
	Sha     string      `json:"sha"`
	Ref     string      `json:"ref"`
	RefType string      `json:"ref_type"`
	Repo    *Repository `json:"repository"`
	Sender  *User       `json:"sender"`
}

// SetSecret FIXME
func (p *CreatePayload) SetSecret(secret string) {
	p.Secret = secret
}

// JSONPayload return payload information
func (p *CreatePayload) JSONPayload() ([]byte, error) {
	return json.MarshalIndent(p, "", "  ")
}

// ParseCreateHook parses create event hook content.
func ParseCreateHook(raw []byte) (*CreatePayload, error) {
	hook := new(CreatePayload)
	if err := json.Unmarshal(raw, hook); err != nil {
		return nil, err
	}

	// it is possible the JSON was parsed, however,
	// was not from Gogs (maybe was from Bitbucket)
	// So we'll check to be sure certain key fields
	// were populated
	switch {
	case hook.Repo == nil:
		return nil, ErrInvalidReceiveHook
	case len(hook.Ref) == 0:
		return nil, ErrInvalidReceiveHook
	}
	return hook, nil
}

// ________         .__          __
// \______ \   ____ |  |   _____/  |_  ____
//  |    |  \_/ __ \|  | _/ __ \   __\/ __ \
//  |    `   \  ___/|  |_\  ___/|  | \  ___/
// /_______  /\___  >____/\___  >__|  \___  >
//         \/     \/          \/          \/

// PusherType define the type to push
type PusherType string

// describe all the PusherTypes
const (
	PusherTypeUser PusherType = "user"
)

// DeletePayload represents delete payload
type DeletePayload struct {
	Ref        string      `json:"ref"`
	RefType    string      `json:"ref_type"`
	PusherType PusherType  `json:"pusher_type"`
	Repo       *Repository `json:"repository"`
	Sender     *User       `json:"sender"`
}

// SetSecret implements Payload
func (p *DeletePayload) SetSecret(secret string) {
}

// JSONPayload implements Payload
func (p *DeletePayload) JSONPayload() ([]byte, error) {
	return json.MarshalIndent(p, "", "  ")
}

// ___________           __
// \_   _____/__________|  | __
//  |    __)/  _ \_  __ \  |/ /
//  |     \(  <_> )  | \/    <
//  \___  / \____/|__|  |__|_ \
//      \/                   \/

// ForkPayload represents fork payload
type ForkPayload struct {
	Forkee *Repository `json:"forkee"`
	Repo   *Repository `json:"repository"`
	Sender *User       `json:"sender"`
}

// SetSecret implements Payload
func (p *ForkPayload) SetSecret(secret string) {
}

// JSONPayload implements Payload
func (p *ForkPayload) JSONPayload() ([]byte, error) {
	return json.MarshalIndent(p, "", "  ")
}

// HookIssueCommentAction defines hook issue comment action
type HookIssueCommentAction string

// all issue comment actions
const (
	HookIssueCommentCreated HookIssueCommentAction = "created"
	HookIssueCommentEdited  HookIssueCommentAction = "edited"
	HookIssueCommentDeleted HookIssueCommentAction = "deleted"
)

// IssueCommentPayload represents a payload information of issue comment event.
type IssueCommentPayload struct {
	Action     HookIssueCommentAction `json:"action"`
	Issue      *Issue                 `json:"issue"`
	Comment    *Comment               `json:"comment"`
	Changes    *ChangesPayload        `json:"changes,omitempty"`
	Repository *Repository            `json:"repository"`
	Sender     *User                  `json:"sender"`
}

// SetSecret implements Payload
func (p *IssueCommentPayload) SetSecret(secret string) {
}

// JSONPayload implements Payload
func (p *IssueCommentPayload) JSONPayload() ([]byte, error) {
	return json.MarshalIndent(p, "", "  ")
}

// __________       .__
// \______   \ ____ |  |   ____ _____    ______ ____
//  |       _// __ \|  | _/ __ \\__  \  /  ___// __ \
//  |    |   \  ___/|  |_\  ___/ / __ \_\___ \\  ___/
//  |____|_  /\___  >____/\___  >____  /____  >\___  >
//         \/     \/          \/     \/     \/     \/

// HookReleaseAction defines hook release action type
type HookReleaseAction string

// all release actions
const (
	HookReleasePublished HookReleaseAction = "published"
)

// ReleasePayload represents a payload information of release event.
type ReleasePayload struct {
	Action     HookReleaseAction `json:"action"`
	Release    *Release          `json:"release"`
	Repository *Repository       `json:"repository"`
	Sender     *User             `json:"sender"`
}

// SetSecret implements Payload
func (p *ReleasePayload) SetSecret(secret string) {
}

// JSONPayload implements Payload
func (p *ReleasePayload) JSONPayload() ([]byte, error) {
	return json.MarshalIndent(p, "", "  ")
}

// __________             .__
// \______   \__ __  _____|  |__
//  |     ___/  |  \/  ___/  |  \
//  |    |   |  |  /\___ \|   Y  \
//  |____|   |____//____  >___|  /
//                      \/     \/

// PushPayload represents a payload information of push event.
type PushPayload struct {
	Secret     string           `json:"secret"`
	Ref        string           `json:"ref"`
	Before     string           `json:"before"`
	After      string           `json:"after"`
	CompareURL string           `json:"compare_url"`
	Commits    []*PayloadCommit `json:"commits"`
	Repo       *Repository      `json:"repository"`
	Pusher     *User            `json:"pusher"`
	Sender     *User            `json:"sender"`
}

// SetSecret FIXME
func (p *PushPayload) SetSecret(secret string) {
	p.Secret = secret
}

// JSONPayload FIXME
func (p *PushPayload) JSONPayload() ([]byte, error) {
	return json.MarshalIndent(p, "", "  ")
}

// ParsePushHook parses push event hook content.
func ParsePushHook(raw []byte) (*PushPayload, error) {
	hook := new(PushPayload)
	if err := json.Unmarshal(raw, hook); err != nil {
		return nil, err
	}

	switch {
	case hook.Repo == nil:
		return nil, ErrInvalidReceiveHook
	case len(hook.Ref) == 0:
		return nil, ErrInvalidReceiveHook
	}
	return hook, nil
}

// Branch returns branch name from a payload
func (p *PushPayload) Branch() string {
	return strings.Replace(p.Ref, "refs/heads/", "", -1)
}

// .___
// |   | ______ ________ __   ____
// |   |/  ___//  ___/  |  \_/ __ \
// |   |\___ \ \___ \|  |  /\  ___/
// |___/____  >____  >____/  \___  >
//          \/     \/            \/

// HookIssueAction FIXME
type HookIssueAction string

const (
	// HookIssueOpened opened
	HookIssueOpened HookIssueAction = "opened"
	// HookIssueClosed closed
	HookIssueClosed HookIssueAction = "closed"
	// HookIssueReOpened reopened
	HookIssueReOpened HookIssueAction = "reopened"
	// HookIssueEdited edited
	HookIssueEdited HookIssueAction = "edited"
	// HookIssueAssigned assigned
	HookIssueAssigned HookIssueAction = "assigned"
	// HookIssueUnassigned unassigned
	HookIssueUnassigned HookIssueAction = "unassigned"
	// HookIssueLabelUpdated label_updated
	HookIssueLabelUpdated HookIssueAction = "label_updated"
	// HookIssueLabelCleared label_cleared
	HookIssueLabelCleared HookIssueAction = "label_cleared"
	// HookIssueSynchronized synchronized
	HookIssueSynchronized HookIssueAction = "synchronized"
	// HookIssueMilestoned is an issue action for when a milestone is set on an issue.
	HookIssueMilestoned HookIssueAction = "milestoned"
	// HookIssueDemilestoned is an issue action for when a milestone is cleared on an issue.
	HookIssueDemilestoned HookIssueAction = "demilestoned"
)

// IssuePayload represents the payload information that is sent along with an issue event.
type IssuePayload struct {
	Secret     string          `json:"secret"`
	Action     HookIssueAction `json:"action"`
	Index      int64           `json:"number"`
	Changes    *ChangesPayload `json:"changes,omitempty"`
	Issue      *Issue          `json:"issue"`
	Repository *Repository     `json:"repository"`
	Sender     *User           `json:"sender"`
}

// SetSecret modifies the secret of the IssuePayload.
func (p *IssuePayload) SetSecret(secret string) {
	p.Secret = secret
}

// JSONPayload encodes the IssuePayload to JSON, with an indentation of two spaces.
func (p *IssuePayload) JSONPayload() ([]byte, error) {
	return json.MarshalIndent(p, "", "  ")
}

// ChangesFromPayload FIXME
type ChangesFromPayload struct {
	From string `json:"from"`
}

// ChangesPayload FIXME
type ChangesPayload struct {
	Title *ChangesFromPayload `json:"title,omitempty"`
	Body  *ChangesFromPayload `json:"body,omitempty"`
}

// __________      .__  .__    __________                                     __
// \______   \__ __|  | |  |   \______   \ ____  ________ __   ____   _______/  |_
//  |     ___/  |  \  | |  |    |       _// __ \/ ____/  |  \_/ __ \ /  ___/\   __\
//  |    |   |  |  /  |_|  |__  |    |   \  ___< <_|  |  |  /\  ___/ \___ \  |  |
//  |____|   |____/|____/____/  |____|_  /\___  >__   |____/  \___  >____  > |__|
//                                     \/     \/   |__|           \/     \/

// PullRequestPayload represents a payload information of pull request event.
type PullRequestPayload struct {
	Secret      string          `json:"secret"`
	Action      HookIssueAction `json:"action"`
	Index       int64           `json:"number"`
	Changes     *ChangesPayload `json:"changes,omitempty"`
	PullRequest *PullRequest    `json:"pull_request"`
	Repository  *Repository     `json:"repository"`
	Sender      *User           `json:"sender"`
}

// SetSecret modifies the secret of the PullRequestPayload.
func (p *PullRequestPayload) SetSecret(secret string) {
	p.Secret = secret
}

// JSONPayload FIXME
func (p *PullRequestPayload) JSONPayload() ([]byte, error) {
	return json.MarshalIndent(p, "", "  ")
}

//__________                           .__  __
//\______   \ ____ ______   ____  _____|__|/  |_  ___________ ___.__.
// |       _// __ \\____ \ /  _ \/  ___/  \   __\/  _ \_  __ <   |  |
// |    |   \  ___/|  |_> >  <_> )___ \|  ||  | (  <_> )  | \/\___  |
// |____|_  /\___  >   __/ \____/____  >__||__|  \____/|__|   / ____|
//        \/     \/|__|              \/                       \/

// HookRepoAction an action that happens to a repo
type HookRepoAction string

const (
	// HookRepoCreated created
	HookRepoCreated HookRepoAction = "created"
	// HookRepoDeleted deleted
	HookRepoDeleted HookRepoAction = "deleted"
)

// RepositoryPayload payload for repository webhooks
type RepositoryPayload struct {
	Secret       string         `json:"secret"`
	Action       HookRepoAction `json:"action"`
	Repository   *Repository    `json:"repository"`
	Organization *User          `json:"organization"`
	Sender       *User          `json:"sender"`
}

// SetSecret set the payload's secret
func (p *RepositoryPayload) SetSecret(secret string) {
	p.Secret = secret
}

// JSONPayload JSON representation of the payload
func (p *RepositoryPayload) JSONPayload() ([]byte, error) {
	return json.MarshalIndent(p, "", " ")
}
