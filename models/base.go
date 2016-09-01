package models

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	_ "time"

	//xormcore "github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	_ "github.com/lib/pq"
	"gitlab.appdao.com/luojie/wechat/conf"
	"gitlab.appdao.com/luojie/wechat/logger"
	"gitlab.appdao.com/luojie/wechat/utils"
)

var (
	dbEngineMaster          = &DBEngine{}
	DBEngineNotAvailableErr = fmt.Errorf("db engine not avaliable")
	dbEngineMux             = sync.RWMutex{}
)

type Session struct {
	*xorm.Session
}

type DBEngine struct {
	*xorm.Engine
	dsn    string
	Status bool
}

func init() {
	rand.Seed(int64(utils.GetNowSecond()))
}

func initDBEngine(engine *DBEngine, config *conf.PGConfig) {
	var err error

	dsn := ""
	if conf.ServiceConfig.PGCluster.Password == "" {
		dsn = fmt.Sprintf(
			"user=%s dbname=%s host=%s port=%d sslmode=disable",
			conf.ServiceConfig.PGCluster.User,
			conf.ServiceConfig.PGCluster.DBName,
			config.Host,
			config.Port)
	} else {
		dsn = fmt.Sprintf(
			"user=%s dbname=%s host=%s port=%d sslmode=disable password=%s",
			conf.ServiceConfig.PGCluster.User,
			conf.ServiceConfig.PGCluster.DBName,
			config.Host,
			config.Port,
			conf.ServiceConfig.PGCluster.Password)
	}

	if engine.Engine, err = xorm.NewEngine("postgres", dsn); err != nil {
		log.Printf("Failed to init db engine: " + err.Error())
		logger.CommonLogger.Fatal(map[string]interface{}{
			"type": "init_db_engine_err",
			"err":  err.Error(),
		})
	}

	engine.dsn = dsn
	engine.SetMaxOpenConns(100)
	engine.SetMaxIdleConns(50)
	//if conf.DebugMode {
	//	engine.Logger().SetLevel(xormcore.LOG_DEBUG)
	//} else {
	//	engine.Logger().SetLevel(xormcore.LOG_DEBUG)
	//}
	//engine.Logger().SetLevel(xormcore.LOG_ERR)
	engine.ShowSQL(conf.ShowSql)

	if err = engine.Ping(); err != nil {
		log.Printf("Failed to ping PostgreSQL <%s>: %s\n", dsn, err.Error())
		logger.CommonLogger.Error(map[string]interface{}{
			"type": "pint_db_engine_err",
			"err":  err.Error(),
		})
		engine.Status = false
	} else {
		engine.Status = true
	}
}

func NewMasterSession() *Session {
	ms := new(Session)
	ms.Session = dbEngineMaster.NewSession()

	return ms
}

func newMasterAutoCloseSession() *Session {
	return newAutoCloseSession(true, false)
}

func GetDBEngine(master, sync bool) (*DBEngine, error) {
	dbEngineMux.RLock()
	defer dbEngineMux.RUnlock()

	if dbEngineMaster.Status {
		return dbEngineMaster, nil
	}

	return nil, DBEngineNotAvailableErr
}

func newAutoCloseSession(master, sync bool) *Session {
	engine, err := GetDBEngine(master, sync)
	if err != nil {
		engine = dbEngineMaster
	}

	ms := new(Session)
	ms.Session = engine.NewSession()
	ms.IsAutoClose = true

	return ms
}

type DBModel interface {
	TableName() string
}

var (
	DBModelDuplicatedErr = errors.New("db model duplicated")
)

func InsertRow(s *Session, m DBModel) (err error) {
	if s == nil {
		s = newMasterAutoCloseSession()
	}
	_, err = s.AllCols().InsertOne(m)

	if err != nil && strings.Index(err.Error(), "duplicate key") >= 0 {
		err = DBModelDuplicatedErr
	}

	return
}

func InsertMultiRows(s *Session, m []interface{}) (err error) {
	var ms *Session

	if s == nil {
		ms = NewMasterSession()
		defer ms.Close()
		if err = ms.Begin(); err != nil {
			return err
		}
	} else {
		ms = s
	}

	_, err = ms.AllCols().Insert(m...)
	if s == nil {
		if err != nil {
			ms.Rollback()
		} else {
			err = ms.Commit()
		}
	}

	if err != nil && strings.Index(err.Error(), "duplicate key") >= 0 {
		err = DBModelDuplicatedErr
	}

	return
}

type UniqueDBModel interface {
	TableName() string
	UniqueCond() (string, []interface{})
}

func UpdateDBModel(s *Session, m UniqueDBModel) (err error) {
	whereStr, whereArgs := m.UniqueCond()
	if s == nil {
		s = newMasterAutoCloseSession()
	}

	_, err = s.AllCols().Where(whereStr, whereArgs...).Update(m)
	if err != nil && strings.Index(err.Error(), "duplicate key") >= 0 {
		err = DBModelDuplicatedErr
	}

	return
}

func DeleteDBModel(s *Session, m UniqueDBModel) (err error) {
	whereStr, whereArgs := m.UniqueCond()

	if s == nil {
		s = newMasterAutoCloseSession()
	}

	_, err = s.Where(whereStr, whereArgs...).Delete(m)
	return
}

func genSequenceValue(sequence string) (int64, error) {
	results, err := dbEngineMaster.Query("select nextval(?) as next", sequence)
	if err != nil {
		return 0, fmt.Errorf("gen %s sequence error: %s", sequence, err.Error())
	}

	id, err := strconv.ParseInt(string(results[0]["next"]), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("gen %s sequence error: %s", sequence, err.Error())
	}

	return id, nil
}
