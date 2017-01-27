package tcpserver

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type Message interface {
	Save(p *Packet) bool                             //存储单条消息
	Range(rid int64, mid int64, limit int) []*Packet //获取大于message id的limit条消息
	SaveMulti(ps []*Packet) bool                     //存储多条消息
}

type MysqlMessage struct {
	db   *sql.DB
	stmt *sql.Stmt
}

func NewMysqlMessage(host string, user string, pwd string, database string, charset string) *MysqlMessage {
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s", user, pwd, host, database, charset)
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	stmt, err := db.Prepare("INSERT INTO `message` (`ver`, `mt`, `mid`, `sid`, `rid`, `ext`, `pl`, `ct`) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return &MysqlMessage{
		db:   db,
		stmt: stmt,
	}
}

func (mm *MysqlMessage) Save(p *Packet) bool {
	//插入数据
	_, err := mm.stmt.Exec(p.Ver, p.Mt, p.Mid, p.Sid, p.Rid, string(p.Ext), string(p.Pl), p.Ct)
	if err != nil {
		fmt.Println(err)
		return false
	}

	return true
}

func (mm *MysqlMessage) SaveMulti(ps []*Packet) bool {
	valueStrings := make([]string, 0, len(ps))
	valueArgs := make([]interface{}, 0, len(ps)*8)
	for _, p := range ps {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs, p.Ver)
		valueArgs = append(valueArgs, p.Mt)
		valueArgs = append(valueArgs, p.Mid)
		valueArgs = append(valueArgs, p.Sid)
		valueArgs = append(valueArgs, p.Rid)
		valueArgs = append(valueArgs, string(p.Ext))
		valueArgs = append(valueArgs, string(p.Pl))
		valueArgs = append(valueArgs, p.Ct)
	}
	stmt := fmt.Sprintf("INSERT INTO `message` (`ver`, `mt`, `mid`, `sid`, `rid`, `ext`, `pl`, `ct`) VALUES %s", strings.Join(valueStrings, ","))
	_, err := mm.db.Exec(stmt, valueArgs...)
	if err != nil {
		fmt.Println(err)
		return false
	}

	return true
}

func (mm *MysqlMessage) Range(rid int64, mid int64, limit int) []*Packet {
	ps := []*Packet{}

	rows, err := mm.db.Query("SELECT `ver`, `mt`, `mid`, `sid`, `rid`, `ext`, `pl`, `ct` FROM `message` WHERE `rid`=? AND `mid`>? ORDER BY mid DESC LIMIT ?", rid, mid, limit)
	if err != nil {
		fmt.Println(err)
		return ps
	}

	for rows.Next() {
		var ver int32
		var mt int32
		var mid int64
		var sid int64
		var rid int64
		var ext string
		var pl string
		var ct int64

		err = rows.Scan(&ver, &mt, &mid, &sid, &rid, &ext, &pl, &ct)
		if err != nil {
			fmt.Println(err)
			continue
		}

		p := &Packet{
			Ver: ver,
			Mt:  mt,
			Mid: mid,
			Sid: sid,
			Rid: rid,
			Ext: []byte(ext),
			Pl:  []byte(pl),
			Ct:  ct,
		}

		ps = append(ps, p)
	}

	rows.Close()
	return ps
}

func (mm *MysqlMessage) Close() {
	if mm.db != nil {
		mm.stmt.Close()
		mm.db.Close()
	}
}
