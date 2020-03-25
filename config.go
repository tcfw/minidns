package main

import (
	"log"

	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault("log_level", "default")

	viper.SetDefault("bind", []string{"127.0.0.1"})
	viper.SetDefault("port", 53)

	viper.SetDefault("disabled_plugins", []string{"forward_resolver"})

	viper.SetDefault("use_internal_resolver", false)

	viper.SetDefault("forwarders", []string{"1.1.1.1", "1.0.0.1"})
	viper.SetDefault("doh_forwarders", []string{"dns.google"})

	viper.SetDefault("blocklists", []string{
		"https://raw.githubusercontent.com/hectorm/hmirror/master/data/adaway.org/list.txt",
		"https://raw.githubusercontent.com/hectorm/hmirror/master/data/adblock-nocoin-list/list.txt",
		"https://raw.githubusercontent.com/hectorm/hmirror/master/data/adguard-simplified/list.txt",
		"https://raw.githubusercontent.com/hectorm/hmirror/master/data/anudeepnd-adservers/list.txt",
		"https://raw.githubusercontent.com/hectorm/hmirror/master/data/disconnect.me-ad/list.txt",
		"https://raw.githubusercontent.com/hectorm/hmirror/master/data/disconnect.me-malvertising/list.txt",
		"https://raw.githubusercontent.com/hectorm/hmirror/master/data/disconnect.me-malware/list.txt",
		"https://raw.githubusercontent.com/hectorm/hmirror/master/data/disconnect.me-tracking/list.txt",
		"https://raw.githubusercontent.com/hectorm/hmirror/master/data/easylist/list.txt",
		"https://raw.githubusercontent.com/hectorm/hmirror/master/data/easyprivacy/list.txt",
		"https://raw.githubusercontent.com/hectorm/hmirror/master/data/eth-phishing-detect/list.txt",
		"https://raw.githubusercontent.com/hectorm/hmirror/master/data/fademind-add.2o7net/list.txt",
		"https://raw.githubusercontent.com/hectorm/hmirror/master/data/fademind-add.dead/list.txt",
		"https://raw.githubusercontent.com/hectorm/hmirror/master/data/fademind-add.risk/list.txt",
		"https://raw.githubusercontent.com/hectorm/hmirror/master/data/fademind-add.spam/list.txt",
		"https://raw.githubusercontent.com/hectorm/hmirror/master/data/kadhosts/list.txt",
		"https://raw.githubusercontent.com/hectorm/hmirror/master/data/malwaredomainlist.com/list.txt",
		"https://raw.githubusercontent.com/hectorm/hmirror/master/data/malwaredomains.com-immortaldomains/list.txt",
		"https://raw.githubusercontent.com/hectorm/hmirror/master/data/malwaredomains.com-justdomains/list.txt",
		"https://raw.githubusercontent.com/hectorm/hmirror/master/data/matomo.org-spammers/list.txt",
		"https://raw.githubusercontent.com/hectorm/hmirror/master/data/mitchellkrogza-badd-boyz-hosts/list.txt",
		"https://raw.githubusercontent.com/hectorm/hmirror/master/data/pgl.yoyo.org/list.txt",
		"https://raw.githubusercontent.com/hectorm/hmirror/master/data/ransomwaretracker.abuse.ch/list.txt",
		"https://raw.githubusercontent.com/hectorm/hmirror/master/data/someonewhocares.org/list.txt",
		"https://raw.githubusercontent.com/hectorm/hmirror/master/data/spam404.com/list.txt",
		"https://raw.githubusercontent.com/hectorm/hmirror/master/data/stevenblack/list.txt",
		"https://raw.githubusercontent.com/hectorm/hmirror/master/data/winhelp2002.mvps.org/list.txt",
		"https://raw.githubusercontent.com/hectorm/hmirror/master/data/zerodot1-coinblockerlists-browser/list.txt",
		"https://raw.githubusercontent.com/hectorm/hmirror/master/data/zeustracker.abuse.ch/list.txt",
		"https://raw.githubusercontent.com/CHEF-KOCH/Audio-fingerprint-pages/master/AudioFp.txt",
		"https://raw.githubusercontent.com/CHEF-KOCH/Canvas-fingerprinting-pages/master/Canvas.txt",
		"https://raw.githubusercontent.com/CHEF-KOCH/WebRTC-tracking/master/WebRTC.txt",
		"https://raw.githubusercontent.com/CHEF-KOCH/CKs-FilterList/master/Anti-Corp/hosts/NSABlocklist.txt",
		"https://gitlab.com/quidsup/notrack-blocklists/raw/master/notrack-blocklist.txt",
		"https://gitlab.com/quidsup/notrack-blocklists/raw/master/notrack-malware.txt",
		"https://www.stopforumspam.com/downloads/toxic_domains_whole.txt",
		"https://dbl.oisd.nl",
		"https://jasonhill.co.uk/pfsense/ytadblock.txt",
	})

	viper.SetEnvPrefix("minidns")
	viper.AutomaticEnv()

	log.Println("Read config")
	log.Printf("Disabled plugins: %v", viper.GetStringSlice("disabled_plugins"))
}
