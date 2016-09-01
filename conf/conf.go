package conf

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/patrickmn/go-cache"
	_ "gitlab.appdao.com/luojie/wechat/utils"
)

var (
	DebugMode bool
	ShowSql   bool

	configFile     = flag.String("config", "__unset__", "service config file")
	maxThreadNum   = flag.Int("max-thread", 0, "max threads of service")
	showSql        = flag.Bool("show-sql", true, "show sql")
	recoveryNotify = flag.String("recovery-notify", "我们的服务正在恢复,您的数据不会丢失,我们马上回来~", "show info to user when server in recovery mode")

	ServiceConfig = &ServiceConfigT{}
	GoCache       = cache.New(cache.NoExpiration, 30*time.Second)
)

const (
	DEFAULT_CACHE_DB_NAME = "default"
)

type PGConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type PGClusterConfig struct {
	User     string    `yaml:"user"`
	Password string    `yaml:"password"`
	DBName   string    `yaml:"db_name"`
	Master   *PGConfig `yaml:"master"`
}
type RedisConfig struct {
	Port   int    `yaml:"port"`
	Host   string `yaml:"host"`
	DBName string `yaml:"db_name"`
}

type WechatMPConfig struct {
	AppId     string `yaml:"appid"`
	AppSecret string `yaml:"appsecret"`
	Token     string `yaml:"token"`
}

type ServiceConfigT struct {
	HttpPort           int              `yaml:"http_port"`
	PGCluster          *PGClusterConfig `yaml:"postgresql_cluster"`
	Redis              *RedisConfig     `yaml:"redis"`
	WeChatMP           *WechatMPConfig  `yaml:"wechatmp"`
	AdminPassword      string           `yaml:"admin_password"`
	LogDir             string           `yaml:"log_dir"`
	RequestLogEnable   bool             `yaml:"request_log_enable"`
	TokenValidDuration int              `yaml:"token_valid_duration"`
}

func init() {
	flag.Parse()

	ShowSql = *showSql

	if len(os.Args) == 2 && os.Args[1] == "reload" {
		wd, _ := os.Getwd()
		pidFile, err := os.Open(filepath.Join(wd, "houdu.pid"))
		if err != nil {
			log.Printf("Failed to open pid file: %s", err.Error())
			os.Exit(1)
		}
		pids := make([]byte, 10)
		n, err := pidFile.Read(pids)
		if err != nil {
			log.Printf("Failed to read pid file: %s", err.Error())
			os.Exit(1)
		}
		if n == 0 {
			log.Printf("No pid in pid file: %s", err.Error())
			os.Exit(1)
		}
		_, err = exec.Command("kill", "-USR2", string(pids[:n])).Output()
		if err != nil {
			log.Printf("Failed to restart service: %s", err.Error())
			os.Exit(1)
		}
		pidFile.Close()
		os.Exit(0)
	}

	if *maxThreadNum == 0 {
		*maxThreadNum = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(*maxThreadNum)

	if *configFile == "__unset__" {
		p, _ := os.Getwd()
		*configFile = filepath.Join(p, "conf/config.yml")
	}

	confFile, err := filepath.Abs(*configFile)
	if err != nil {
		log.Printf("No correct config file: %s - %s", *configFile, err.Error())
		os.Exit(1)
	}

	confBs, err := ioutil.ReadFile(confFile)
	if err != nil {
		log.Printf("Failed to read config fliel <%s> : %s", confFile, err.Error())
		os.Exit(1)
	}

	err = yaml.Unmarshal(confBs, ServiceConfig)
	if err != nil {
		log.Printf("Failed to parse config fliel <%s> : %s", confFile, err.Error())
		os.Exit(1)
	}

}
