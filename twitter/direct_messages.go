package twitter

import (
	"net/http"
	"strconv"
	"time"

	"github.com/dghubble/sling"
)

// DirectMessageEvents lists Direct Message events.
type (
	DirectMessageEvents struct {
		Events     []DirectMessageEvent `json:"events"`
		NextCursor string               `json:"next_cursor"`
	}

	// DirectMessageEvent is a single Direct Message sent or received.
	DirectMessageEvent struct {
		CreatedAt string                     `json:"created_timestamp"`
		ID        string                     `json:"id"`
		Type      string                     `json:"type"`
		Message   *DirectMessageEventMessage `json:"message_create"`
	}

	// DirectMessageEventMessage contains message contents, along with sender and
	// target recipient.
	DirectMessageEventMessage struct {
		SenderID         string               `json:"sender_id"`
		SenderScreenName string               `json:"sender_screen_name"`
		Target           DMEventMessageTarget `json:"target"`
		Data             DMEventMessageData   `json:"message_data"`
	}

	DMEventMessageTarget struct {
		RecipientID         string `json:"recipient_id"`
		RecipientScreenName string `json:"recipient_screen_name"`
	}

	DMEventMessageData struct {
		Text       string                       `json:"text"`
		Entities   *Entities                    `json:"entities"`
		Attachment DMEventMessageDataAttachment `json:"attachment"`
	}

	DMEventMessageDataAttachment struct {
		Type  string      `json:"type"`
		Media MediaEntity `json:"media"`
	}

	// DirectMessageService provides methods for accessing Twitter direct message
	// API endpoints.
	DirectMessageService struct {
		baseSling *sling.Sling
		sling     *sling.Sling
	}

	// DirectMessageEventsCreateResponse represents the data structure returned by
	// the Twitter API after the successful creation of an event
	DirectMessageEventsCreateResponse struct {
		Event DirectMessageEvent
	}

	// DirectMessageEventsListParams are the parameters for
	// DirectMessageService.EventsList
	DirectMessageEventsListParams struct {
		Cursor string `url:"cursor,omitempty"`
		Count  int    `url:"count,omitempty"`
	}

	directMessageEventsCreateData struct {
		Text string `json:"text"`
	}
	directMessageEventsCreateTarget struct {
		RecipientID string `json:"recipient_id"`
	}
	directMessageEventsCreateMessage struct {
		Target *directMessageEventsCreateTarget `json:"target"`
		Data   *directMessageEventsCreateData   `json:"message_data"`
	}
	directMessageEventsCreateEvent struct {
		Type    string                            `json:"type"`
		Message *directMessageEventsCreateMessage `json:"message_create"`
	}
	directMessageEventsCreateParams struct {
		Event *directMessageEventsCreateEvent `json:"event"`
	}
)

// newDirectMessageService returns a new DirectMessageService.
func newDirectMessageService(sling *sling.Sling) *DirectMessageService {
	return &DirectMessageService{
		baseSling: sling.New(),
		sling:     sling.Path("direct_messages/"),
	}
}

// EventsList returns Direct Message events (both sent and received) within
// the last 30 days in reverse chronological order.
// Requires a user auth context with DM scope.
// https://developer.twitter.com/en/docs/direct-messages/sending-and-receiving/api-reference/list-events
func (s *DirectMessageService) EventsList(params *DirectMessageEventsListParams) (*DirectMessageEvents, *http.Response, error) {
	events := new(DirectMessageEvents)
	apiError := new(APIError)
	resp, err := s.sling.New().Get("events/list.json").QueryStruct(params).Receive(events, apiError)
	return events, resp, relevantError(err, *apiError)
}

// EventsCreate creates a new Direct Message event
func (s *DirectMessageService) EventsCreate(event *DirectMessageEventMessage) (*DirectMessageEventsCreateResponse, *http.Response, error) {
	apiParams := &directMessageEventsCreateParams{
		Event: &directMessageEventsCreateEvent{
			Type: "message_create",
			Message: &directMessageEventsCreateMessage{
				Target: &directMessageEventsCreateTarget{
					RecipientID: event.Target.RecipientID,
				},
				Data: &directMessageEventsCreateData{
					Text: event.Data.Text,
				},
			},
		},
	}
	apiError := new(APIError)
	eventResponse := new(DirectMessageEventsCreateResponse)
	resp, err := s.sling.New().Post("events/new.json").BodyJSON(apiParams).Receive(eventResponse, apiError)
	return eventResponse, resp, relevantError(err, *apiError)
}

// CreatedAtTime returns the UTC time a Direct Message was created from Twitter's timestamp.
func (d DirectMessageEvent) CreatedAtTime() (date time.Time, err error) {
	milli, err := strconv.ParseInt(d.CreatedAt, 10, 64)
	if err != nil {
		return date, err
	}
	date = time.Unix(milli/1000, 0).UTC()
	return date, err
}

// DEPRECATED ********************************************************************************************************************************************

// DirectMessage is a direct message to a single recipient (DEPRECATED).
type DirectMessage struct {
	CreatedAt           string    `json:"created_at"`
	Entities            *Entities `json:"entities"`
	ID                  int64     `json:"id"`
	IDStr               string    `json:"id_str"`
	Recipient           *User     `json:"recipient"`
	RecipientID         int64     `json:"recipient_id"`
	RecipientScreenName string    `json:"recipient_screen_name"`
	Sender              *User     `json:"sender"`
	SenderID            int64     `json:"sender_id"`
	SenderScreenName    string    `json:"sender_screen_name"`
	Text                string    `json:"text"`
}

// CreatedAtTime returns the time a Direct Message was created (DEPRECATED).
func (d DirectMessage) CreatedAtTime() (time.Time, error) {
	return time.Parse(time.RubyDate, d.CreatedAt)
}

// directMessageShowParams are the parameters for DirectMessageService.Show
type directMessageShowParams struct {
	ID int64 `url:"id,omitempty"`
}

// Show returns the requested Direct Message (DEPRECATED).
// Requires a user auth context with DM scope.
// https://dev.twitter.com/rest/reference/get/direct_messages/show
func (s *DirectMessageService) Show(id int64) (*DirectMessage, *http.Response, error) {
	params := &directMessageShowParams{ID: id}
	dm := new(DirectMessage)
	apiError := new(APIError)
	resp, err := s.sling.New().Get("show.json").QueryStruct(params).Receive(dm, apiError)
	return dm, resp, relevantError(err, *apiError)
}

// DirectMessageGetParams are the parameters for DirectMessageService.Get
// (DEPRECATED).
type DirectMessageGetParams struct {
	SinceID         int64 `url:"since_id,omitempty"`
	MaxID           int64 `url:"max_id,omitempty"`
	Count           int   `url:"count,omitempty"`
	IncludeEntities *bool `url:"include_entities,omitempty"`
	SkipStatus      *bool `url:"skip_status,omitempty"`
}

// Get returns recent Direct Messages received by the authenticated user
// (DEPRECATED).
// Requires a user auth context with DM scope.
// https://dev.twitter.com/rest/reference/get/direct_messages
func (s *DirectMessageService) Get(params *DirectMessageGetParams) ([]DirectMessage, *http.Response, error) {
	dms := new([]DirectMessage)
	apiError := new(APIError)
	resp, err := s.baseSling.New().Get("direct_messages.json").QueryStruct(params).Receive(dms, apiError)
	return *dms, resp, relevantError(err, *apiError)
}

// DirectMessageSentParams are the parameters for DirectMessageService.Sent
// (DEPRECATED).
type DirectMessageSentParams struct {
	SinceID         int64 `url:"since_id,omitempty"`
	MaxID           int64 `url:"max_id,omitempty"`
	Count           int   `url:"count,omitempty"`
	Page            int   `url:"page,omitempty"`
	IncludeEntities *bool `url:"include_entities,omitempty"`
}

// Sent returns recent Direct Messages sent by the authenticated user
// (DEPRECATED).
// Requires a user auth context with DM scope.
// https://dev.twitter.com/rest/reference/get/direct_messages/sent
func (s *DirectMessageService) Sent(params *DirectMessageSentParams) ([]DirectMessage, *http.Response, error) {
	dms := new([]DirectMessage)
	apiError := new(APIError)
	resp, err := s.sling.New().Get("sent.json").QueryStruct(params).Receive(dms, apiError)
	return *dms, resp, relevantError(err, *apiError)
}

// DirectMessageNewParams are the parameters for DirectMessageService.New
// (DEPRECATED).
type DirectMessageNewParams struct {
	UserID     int64  `url:"user_id,omitempty"`
	ScreenName string `url:"screen_name,omitempty"`
	Text       string `url:"text"`
}

// New sends a new Direct Message to a specified user as the authenticated
// user (DEPRECATED).
// Requires a user auth context with DM scope.
// https://dev.twitter.com/rest/reference/post/direct_messages/new
func (s *DirectMessageService) New(params *DirectMessageNewParams) (*DirectMessage, *http.Response, error) {
	dm := new(DirectMessage)
	apiError := new(APIError)
	resp, err := s.sling.New().Post("new.json").BodyForm(params).Receive(dm, apiError)
	return dm, resp, relevantError(err, *apiError)
}

// DirectMessageDestroyParams are the parameters for DirectMessageService.Destroy
// (DEPRECATED).
type DirectMessageDestroyParams struct {
	ID              int64 `url:"id,omitempty"`
	IncludeEntities *bool `url:"include_entities,omitempty"`
}

// Destroy deletes the Direct Message with the given id and returns it if
// successful (DEPRECATED).
// Requires a user auth context with DM scope.
// https://dev.twitter.com/rest/reference/post/direct_messages/destroy
func (s *DirectMessageService) Destroy(id int64, params *DirectMessageDestroyParams) (*DirectMessage, *http.Response, error) {
	if params == nil {
		params = &DirectMessageDestroyParams{}
	}
	params.ID = id
	dm := new(DirectMessage)
	apiError := new(APIError)
	resp, err := s.sling.New().Post("destroy.json").BodyForm(params).Receive(dm, apiError)
	return dm, resp, relevantError(err, *apiError)
}
