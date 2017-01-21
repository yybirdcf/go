package models

type ModeGame struct {
	Id             uint   `gorm:"column:id"`
	Username       string `gorm:"column:username"`
	Phone          string `gorm:"column:phone"`
	Sex            int    `gorm:"column:sex"`
	Createtime     int    `gorm:"column:createtime"`
	Status         int    `gorm:"column:status"`
	Avatar         string `gorm:"column:avatar"`
	Gouhao         int    `gorm:"column:gouhao"`
	Birthday       int    `gorm:"column:birthday"`
	UpdateTime     int    `gorm:"column:update_time"`
	Avatars        string `gorm:"column:avatars"`
	LastLogin      int    `gorm:"column:last_login"`
	Signature      string `gorm:"column:signature"`
	Appfrom        string `gorm:"column:appfrom"`
	Appver         string `gorm:"column:appver"`
	BackgroudImage string `gorm:"column:backgroud_image"`
	UpdateAppver   string `gorm:"column:update_appver"`
	AccessToken    string `gorm:"column:access_token"`
	LoginPwd       string `gorm:"column:login_pwd"`
	Privacy        int    `gorm:"column:privacy"`
	LoadRecTags    int    `gorm:"column:loadRecTags"`
	GamePower      int    `gorm:"column:game_power"`
	Mark           int    `gorm:"column:mark"`
	GreetWord      string `gorm:"column:greet_word"`
	GreetWordFirst int    `gorm:"column:greet_word_first"`
	Invalid        int    `gorm:"column:invalid"`
	Level          int    `gorm:"column:level"`
	QuestionPhoto  string `gorm:"column:question_photo"`
	Lan            string `gorm:"column:lan"`
	Notify         int    `gorm:"column:notify"`
	AppfromOri     string `gorm:"column:appfrom_ori"`
	Appid          string `gorm:"column:appid"`
}

func (mg ModeGame) TableName() string {
	return "modegame"
}
