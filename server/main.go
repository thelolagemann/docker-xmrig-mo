// package main provides a small http server that serves the
// xmrig-workers UI, as well as proxying any requests
// made to http://<ip>:3001/api/$1 => http://<ip>:3000/$1.
// This is to prevent confusion and misconfiguration between
// the ports used to communicate with the xmrig API.
package main

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	configLocation string = "/cfg/config.json"
)

var (
	accessToken string
)

func main() {
	// bootstrap config
	c, err := openConfig()
	if err != nil {
		panic(err)
	}
	for env, f := range configMap {
		envs := strings.Split(env, ",")
		for _, e := range envs {
			if os.Getenv(e) != "" {
				f(c)
			}
		}
	}

	// configure rig name
	var rigName string
	if os.Getenv("RIG_NAME") != "" {
		rigName = os.Getenv("RIG_NAME")
	} else {
		rigName = randomString(8)
	}
	c.Pools[0].Pass = rigName

	// save config
	if err := saveConfig(c); err != nil {
		panic(err)
	}

	// http handling
	http.HandleFunc("/", handleRequestOrRedirect)
	if err := http.ListenAndServe(":3001", nil); err != nil {
		panic(err)
	}
}

func handleRequestOrRedirect(res http.ResponseWriter, req *http.Request) {
	target, _ := url.Parse("http://localhost:3000")

	if req.Header.Get("Authorization") != "" {
		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.ServeHTTP(res, req)
	} else {
		if os.Getenv("XMRIG_WORKERS_AUTOCONFIGURE") == "true" {
			if req.URL.Path == "/" || req.URL.Path == "/index.html" {
				f, err := os.OpenFile("/xmrig-workers/www/index.html", os.O_RDONLY, 0755)
				if err != nil {
					panic(err)
				}
				b, err := io.ReadAll(f)
				if err != nil {
					panic(err)
				}
				new := strings.Replace(string(b), "</body>", `<script>window.localStorage.setItem("xmrig.workers", "[[\"" + window.location.href.replace("/#/", "") + "\",\"`+accessToken+`\",\"\", \"v6.14.1-mo2\","+Date.now()+"]]")</script></body>`, -1)
				res.Write([]byte(new))
				res.WriteHeader(200)
				return
			}
		}
		http.FileServer(http.Dir("/xmrig-workers/www")).ServeHTTP(res, req)
	}
}

var (
	configMap = map[string]func(c *config){
		"XMRIG_API_ENABLED,XMRIG_WORKERS_ENABLED": func(c *config) {
			if os.Getenv("API_TOKEN") != "" {
				accessToken = os.Getenv("API_TOKEN")
			} else {
				accessToken = randomString(16)
			}
			c.HTTP.AccessToken = accessToken
		},
		"BENCHMARK": func(c *config) {
			bTime, err := strconv.Atoi(os.Getenv("BENCHMARK_TIME"))
			if err != nil {
				panic(err)
			}
			c.BenchAlgoTime = bTime
			c.RebenchAlgo = true
		},
		"WALLET_ADDRESS": func(c *config) {
			c.Pools[0].User = os.Getenv("WALLET_ADDRESS")
		},
	}
)

type config struct {
	API struct {
		ID       interface{} `json:"id"`
		WorkerID interface{} `json:"worker-id"`
	} `json:"api"`
	HTTP struct {
		Enabled     bool   `json:"enabled"`
		Host        string `json:"host"`
		Port        int    `json:"port"`
		AccessToken string `json:"access-token"`
		Restricted  bool   `json:"restricted"`
	} `json:"http"`
	Autosave   bool `json:"autosave"`
	Background bool `json:"background"`
	Colors     bool `json:"colors"`
	Title      bool `json:"title"`
	Randomx    struct {
		Init                   int    `json:"init"`
		InitAvx2               int    `json:"init-avx2"`
		Mode                   string `json:"mode"`
		OneGbPages             bool   `json:"1gb-pages"`
		Rdmsr                  bool   `json:"rdmsr"`
		Wrmsr                  bool   `json:"wrmsr"`
		CacheQos               bool   `json:"cache_qos"`
		Numa                   bool   `json:"numa"`
		ScratchpadPrefetchMode int    `json:"scratchpad_prefetch_mode"`
	} `json:"randomx"`
	CPU struct {
		Enabled         bool        `json:"enabled"`
		HugePages       bool        `json:"huge-pages"`
		HugePagesJit    bool        `json:"huge-pages-jit"`
		HwAes           interface{} `json:"hw-aes"`
		Priority        interface{} `json:"priority"`
		MemoryPool      bool        `json:"memory-pool"`
		Yield           bool        `json:"yield"`
		Asm             bool        `json:"asm"`
		Argon2Impl      interface{} `json:"argon2-impl"`
		AstrobwtMaxSize int         `json:"astrobwt-max-size"`
		AstrobwtAvx2    bool        `json:"astrobwt-avx2"`
		Argon2          []int       `json:"argon2"`
		Astrobwt        []int       `json:"astrobwt"`
		Cn              [][]int     `json:"cn"`
		CnHeavy         [][]int     `json:"cn-heavy"`
		CnLite          [][]int     `json:"cn-lite"`
		CnPico          [][]int     `json:"cn-pico"`
		CnGpu           [][]int     `json:"cn/gpu"`
		CnUpx2          [][]int     `json:"cn/upx2"`
		Panthera        []int       `json:"panthera"`
		Rx              []int       `json:"rx"`
		RxWow           []int       `json:"rx/wow"`
		CnLite0         bool        `json:"cn-lite/0"`
		Cn0             bool        `json:"cn/0"`
		RxArq           string      `json:"rx/arq"`
		RxKeva          string      `json:"rx/keva"`
	} `json:"cpu"`
	LogFile         interface{} `json:"log-file"`
	DonateLevel     int         `json:"donate-level"`
	DonateOverProxy int         `json:"donate-over-proxy"`
	Pools           []struct {
		Algo           interface{} `json:"algo"`
		Coin           interface{} `json:"coin"`
		URL            string      `json:"url"`
		User           string      `json:"user"`
		Pass           string      `json:"pass"`
		RigID          interface{} `json:"rig-id"`
		Nicehash       bool        `json:"nicehash"`
		Keepalive      bool        `json:"keepalive"`
		Enabled        bool        `json:"enabled"`
		TLS            bool        `json:"tls"`
		TLSFingerprint interface{} `json:"tls-fingerprint"`
		Daemon         bool        `json:"daemon"`
		Socks5         interface{} `json:"socks5"`
		SelfSelect     interface{} `json:"self-select"`
		SubmitToOrigin bool        `json:"submit-to-origin"`
	} `json:"pools"`
	Retries    int  `json:"retries"`
	RetryPause int  `json:"retry-pause"`
	PrintTime  int  `json:"print-time"`
	Dmi        bool `json:"dmi"`
	Syslog     bool `json:"syslog"`
	TLS        struct {
		Enabled      bool        `json:"enabled"`
		Protocols    interface{} `json:"protocols"`
		Cert         interface{} `json:"cert"`
		CertKey      interface{} `json:"cert_key"`
		Ciphers      interface{} `json:"ciphers"`
		Ciphersuites interface{} `json:"ciphersuites"`
		Dhparam      interface{} `json:"dhparam"`
	} `json:"tls"`
	DNS struct {
		Ipv6 bool `json:"ipv6"`
		TTL  int  `json:"ttl"`
	} `json:"dns"`
	UserAgent     interface{} `json:"user-agent"`
	Verbose       int         `json:"verbose"`
	Watch         bool        `json:"watch"`
	RebenchAlgo   bool        `json:"rebench-algo"`
	BenchAlgoTime int         `json:"bench-algo-time"`
	AlgoPerf      struct {
		Cn1            float64 `json:"cn/1"`
		Cn2            float64 `json:"cn/2"`
		CnR            float64 `json:"cn/r"`
		CnFast         float64 `json:"cn/fast"`
		CnHalf         float64 `json:"cn/half"`
		CnXao          float64 `json:"cn/xao"`
		CnRto          float64 `json:"cn/rto"`
		CnRwz          float64 `json:"cn/rwz"`
		CnZls          float64 `json:"cn/zls"`
		CnDouble       float64 `json:"cn/double"`
		CnCcx          int     `json:"cn/ccx"`
		CnLite1        float64 `json:"cn-lite/1"`
		CnHeavyXhv     float64 `json:"cn-heavy/xhv"`
		CnPico         float64 `json:"cn-pico"`
		CnPicoTlo      float64 `json:"cn-pico/tlo"`
		CnGpu          float64 `json:"cn/gpu"`
		Panthera       float64 `json:"panthera"`
		Rx0            float64 `json:"rx/0"`
		RxWow          float64 `json:"rx/wow"`
		RxArq          float64 `json:"rx/arq"`
		RxSfx          float64 `json:"rx/sfx"`
		Argon2Chukwav2 float64 `json:"argon2/chukwav2"`
		Astrobwt       int     `json:"astrobwt"`
	} `json:"algo-perf"`
	PauseOnBattery bool `json:"pause-on-battery"`
	PauseOnActive  bool `json:"pause-on-active"`
}

func openConfig() (*config, error){
	f, err := os.Open(configLocation)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var cfg *config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	return cfg, f.Close()
}

func saveConfig(c *config) error {
	f, err := os.OpenFile(configLocation, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	if b, err := json.MarshalIndent(c, "", "\t"); err != nil {
		return err
	} else {
		f.Write(b)
	}

	return f.Close()
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)
var src = rand.NewSource(time.Now().UnixNano())

func randomString(n int) string {
	sb := strings.Builder{}
	sb.Grow(n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return sb.String()
}

