package notify

const (
	Port        = "android/notify"
	PortChannel = "android/notify/channel"
)

type Notify chan<- []Notification

type Notification struct {
	Id            int
	ChannelId     string
	ContentTitle  string
	ContentText   string
	ContentInfo   string
	Ticker        string
	SubText       string
	SmallIcon     string
	Number        int
	Ongoing       bool
	Silent        bool
	Defaults      int
	OnlyAlertOnce bool
	AutoCancel    bool
	Group         string
	GroupSummary  bool
	Priority      int
	ContentIntent *Intent
	Progress      *Progress
	Action        *Action
}

type Progress struct {
	Max           int
	Current       int
	Indeterminate bool
}

type Intent struct {
	Type   string
	Action string
	Uri    string
}

type Channel struct {
	Id         string
	Name       string
	Importance int
}

type Action struct {
	Icon   string
	Title  string
	Intent *Intent
}

const (
	// ImportanceUnspecified signifying that the user has not expressed an importance. This value is for persisting preferences, and should never be associated with an actual notification.
	ImportanceUnspecified = -1000
	// ImportanceNone a notification with no importance: shows nowhere, is blocked.
	ImportanceNone = 0
	// ImportanceMin min notification importance: only shows in the shade, below the fold.
	ImportanceMin = 1
	// ImportanceLow low notification importance: shows everywhere, but is not intrusive.
	ImportanceLow = 2
	// ImportanceDefault default notification importance: shows everywhere, allowed to makes noise, but does not visually intrude.
	ImportanceDefault = 3
	// ImportanceHigh higher notification importance: shows everywhere, allowed to makes noise and peek.
	ImportanceHigh = 4
	// ImportanceMax the highest notification importance: shows everywhere, allowed to makes noise, peek, and use full screen intents.
	ImportanceMax = 5
)

const (
	// PriorityDefault Default notification priority. If your application does not prioritize its own notifications, use this value for all notifications.
	PriorityDefault = 0
	// PriorityLow Lower notification priority, for items that are less important. The UI may choose to show these items smaller, or at a different position in the list, compared with your app's {@link #PRIORITY_DEFAULT} items.
	PriorityLow = -1
	// PriorityMin the lowest notification priority. These items might not be shown to the user except under special circumstances, such as detailed notification logs.
	PriorityMin = -2
	// PriorityHigh higher notification priority, for more important notifications or alerts. The UI may choose to show these items larger, or at a different position in notification lists, compared with your app's {@link #PRIORITY_DEFAULT} items.
	PriorityHigh = 1
	// PriorityMax the highest notification priority, for your application's most important items that require the user's prompt attention or input.
	PriorityMax = 2
)
