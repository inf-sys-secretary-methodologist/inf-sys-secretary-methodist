package dto

// CalendarSubscriptionOutput describes a user's calendar feed subscription.
// The URL is the secret iCalendar address to paste into an external calendar;
// it is empty when the user has no active subscription.
type CalendarSubscriptionOutput struct {
	Subscribed bool   `json:"subscribed"`
	URL        string `json:"url,omitempty"`
}
