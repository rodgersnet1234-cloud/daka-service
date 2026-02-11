package model

import (
	"database/sql"

	"github.com/gorilla/websocket"
)

// Client represents a connected user
type Client struct {
	Phone string
	Conn  *websocket.Conn
}

// AuthMessage is sent by client immediately after connecting
type AuthMessage struct {
	Type  string `json:"type"`
	Phone string `json:"phone"`
}

// NotifyMessage sent from server to client
type NotifyMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type LevelThreshold struct {
	Level     string
	Threshold int
}

type ErrMsgs struct {
	ID        int
	ErrorName string
}

type PackageLog struct {
	ID           int
	UserID       int
	PackageID    int
	Amount       string // to be parsed as float64 when needed
	AddTime      string // can use time.Time if parsed
	Frozen       int
	Active       int
	Day          int
	AlreadyAdded string
	Method       string
	Days         int
}
type Package struct {
	ID             int
	TradableAmount float64
	DailyReward    float64
	PackageValue   string
	TotalReward    float64
}
type PackageTask struct {
	Task   string `json:"task"`
	Amount string `json:"amount"`
	ID     int    `json:"id"`
}
type RUser struct {
	ID            int
	Name          string
	Phone         string
	PackageValue  int
	PackageLevel  int
	Packages      int
	Level         int
	DownlineCount int
	Invited       sql.NullInt64
	AV            string
	GroupAV       string
	Shouyi        string
	Jifen         string
}

type Shouyi struct {
	ID            string `json:"id"`
	UserID        string `json:"user_id"`
	Source        string `json:"source"`
	Amount        string `json:"amount"`
	Time          string `json:"time"`
	Packages      string `json:"packages"`
	Trans         string `json:"trans"`
	BalanceBefore string `json:"balance_before"`
	BalanceAfter  string `json:"balance_after"`
	Type          string `json:"type"`
	PackageID     string `json:"package_id"`
}
