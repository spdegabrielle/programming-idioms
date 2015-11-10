package pigae

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"appengine"
)

// Bad stuff started 2015-10...
// Had to take action.

var spammerSet = map[string]bool{}

func init() {
	b1, err := ioutil.ReadFile("spammers.csv")
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	b2, err := ioutil.ReadFile("spammers-custom.csv")
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	spammerSet = make(map[string]bool)
	for _, b := range [][]byte{b1, b2} {
		full := string(b)
		spammers := strings.Split(full, ",")
		for _, spammer := range spammers {
			spammerSet[spammer] = true
		}
	}
	// How to log this at startup ...?
	// fmt.Println("Registered list of", len(spammerSet), "spam IPs.")
}

func isSpam(w http.ResponseWriter, r *http.Request) (busted bool) {
	var motive string
	ip := r.RemoteAddr

	defer func() {
		if busted {
			c := appengine.NewContext(r)
			c.Infof("Detected spammer %v : %v", ip, motive)
			// Let's return a nice 200 ... nothing to see here
			fmt.Fprintln(w, "<html><body>This site is under construction</body></html>")
		}
	}()

	// By suspicious IP
	if spammerSet[ip] {
		motive = "blacklisted"
		return true
	}

	// By suspicious field values
	if len(r.FormValue("impl_imports")) > 150 {
		motive = "imports too long [" + r.FormValue("impl_imports")[0:20] + "...]"
		return true
	}
	if len(r.FormValue("impl_imports")) > 10 && r.FormValue("impl_imports") == r.FormValue("impl_code") {
		motive = "identical imports and codeblock [" + r.FormValue("impl_imports")[0:10] + "...]"
		return true
	}

	return false
}
