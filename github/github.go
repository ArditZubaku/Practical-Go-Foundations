package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	fmt.Println(gitHubInfo(ctx, "ArditZubaku"))
	defer cancel()

	// JSON: io.Reader -> Go => json.Decoder
	// JSON: []byte -> Go => json.Unmarshal
	// Go: io.Writer -> JSON => json.Encoder
	// Go: []byte -> JSON => json.Marshal
}

func gitHubInfo(ctx context.Context, login string) (string, int, error) {
	// url := fmt.Sprintf("https://api.github.com/users/%s", login)
	url := "https://api.github.com/users/" + url.PathEscape(login)
	// resp, err := http.Get(url)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		// log.Fatalf("ERROR: %s", err)
		return "", 0, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// log.Fatalf("ERROR: %s", err)
		return "", 0, err
	}

	if resp.StatusCode != http.StatusOK {
		// log.Fatalf("ERROR: %s", resp.Status)
		return "", 0, fmt.Errorf("%#v - %s", url, resp.Status)
	}

	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))

	/*
		* Once you read from the body you have basically consumed it, it is an io.Reader
			if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
				log.Fatalf("ERROR: Can't copy - %s\n", err)
			}
	*/

	// var r Reply

	// Anonymous struct
	var r struct {
		Name        string `json:"name"`
		PublicRepos int    `json:"public_repos"`
	}

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&r); err != nil {
		// log.Fatalf("ERROR: Can't decode - %s", err)
		return "", 0, err
	}

	defer resp.Body.Close()

	// fmt.Println(r)
	// This is the best
	fmt.Printf("%#v", r)
	// fmt.Printf("Extra verbose? > %+v", r)

	return r.Name, r.PublicRepos, nil
}

// Fields we want to be (Un)Marshaled should be exported
type Reply struct {
	Name        string `json:"name"`
	PublicRepos int    `json:"public_repos"`
}
