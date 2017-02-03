package pushserver

import (
  "time"
  "net/http"
  "encoding/json"
  "strconv"
)

//apns请求头信息
type Headers struct {
  ID string //通知id
  CollapseID string //通知id，用来更新一个已经存在的通知
  Expiration time.Time //过期时间，apple在过期时间内会重试，默认只会发送一次
  LowPriority bool //是否允许apple将消息合并一起发送以减少电池消耗，默认立即推送
  Topic string //bundle id包名
}

//设置http请求头信息
func (h *Headers) set(reqHeader *http.Header) {
  if h == nil {
    return
  }

  //推送消息可以忽略，苹果会生成uuid
  if h.ID != "" {
    reqHeader.Set("apns-id", h.ID)
  }

  if h.CollapseID != "" {
    reqHeader.Set("apns-collapse-id", h.CollapseID)
  }

  if !h.Expiration.IsZero() {
    reqHeader.Set("apns-expiration", strconv.FormatInt(h.Expiration.Unix(), 10))
  }

  //忽略，默认是10
  if h.LowPriority {
    reqHeader.Set("apns-priority", "5")
  }

  if h.Topic != "" {
    reqHeader.Set("apns-topic", h.Topic)
  }
}

//通知消息结构体
type Alert struct {
  Title string `json:"title,omitempty"`
  TitleLocKey  string   `json:"title-loc-key,omitempty"`
	TitleLocArgs []string `json:"title-loc-args,omitempty"`

	//子标题 iOS 10生效
	Subtitle string `json:"subtitle,omitempty"`

	//消息体
	Body    string   `json:"body,omitempty"`
  //Key to an alert-message string in a Localizable
	LocKey  string   `json:"loc-key,omitempty"`
  //Variable string values to appear in place of the format specifiers in locKey
	LocArgs []string `json:"loc-args,omitempty"`

	//If a value is specified for the actionLocKey argument, an alert with two buttons is displayed. The value is a key to get a localized string in a Localizable.strings file to use for the right button’s title
	ActionLocKey string `json:"action-loc-key,omitempty"`

	//启动应用图标
	LaunchImage string `json:"launch-image,omitempty"`
}

// isSimple alert with only Body set.
func (a *Alert) isSimple() bool {
	return len(a.Title) == 0 && len(a.Subtitle) == 0 &&
		len(a.LaunchImage) == 0 &&
		len(a.TitleLocKey) == 0 && len(a.TitleLocArgs) == 0 &&
		len(a.LocKey) == 0 && len(a.LocArgs) == 0 && len(a.ActionLocKey) == 0
}

// isZero if no Alert fields are set.
func (a *Alert) isZero() bool {
	return len(a.Body) == 0 && a.isSimple()
}

type Badge struct {
  Number uint
  Isset bool
}

func (b *Badge) SetBadge(n uint) {
  b.Isset = true
  b.Number = n
}

func (b *Badge) IsSet() bool {
  return b.Isset
}

func (b *Badge) N() uint {
  return b.Number
}

//payload aps设置
type APS struct {
  Alert Alert
  //app应用图标角标数字
  Badge Badge
  //声音
  Sound string
  //Content available is for silent notifications
	// with no alert, sound, or badge.
  ContentAvailable bool
  // Category identifier for custom actions in iOS 8 or newer
  Category string
  // Mutable is used for Service Extensions introduced in iOS 10.
  MutableContent bool
}

func (a *APS) Map() map[string]interface{} {
  aps := make(map[string]interface{})

  if !a.Alert.isZero() {
		if a.Alert.isSimple() {
			aps["alert"] = a.Alert.Body
		} else {
			aps["alert"] = a.Alert
		}
	}

  if a.Badge.IsSet() {
    aps["badge"] = a.Badge.N()
  }

  if a.Sound != "" {
		aps["sound"] = a.Sound
	}
	if a.ContentAvailable {
		aps["content-available"] = 1
	}
	if a.Category != "" {
		aps["category"] = a.Category
	}
	if a.MutableContent {
		aps["mutable-content"] = 1
	}

	return map[string]interface{}{"aps": aps}
}

func (a *APS) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Map())
}
