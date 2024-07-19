package gitlab

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/samber/lo"
)

type Client struct {
	host      string
	authToken string
}

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

type References struct {
	Short string `json:"short"`
	Full  string `json:"full"`
}

type MR struct {
	ID           int64      `json:"id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Reviewers    []User     `json:"reviewers"`
	Assignees    []User     `json:"assignees"`
	WebUrl       string     `json:"web_url"`
	Author       User       `json:"author"`
	References   References `json:"references"`
	ProjectID    int64      `json:"project_id"`
	IID          int64      `json:"iid"`
	ApprovedBy   []User
	ApprovedByMe bool
}

type Note struct {
	Body   string `json:"body"`
	Author User   `json:"author"`
}

const approveNoteBody = "approved this merge request"

var userMentionRegex = regexp.MustCompile("@[a-z.]*")

func NewClient(host, authToken string) *Client {
	return &Client{
		host:      host,
		authToken: authToken,
	}
}

func (c *Client) CheckAuth() error {
	return c.Do(http.MethodGet, "/api/v4/user", "")
}

func (c *Client) GetCurrentUsername() (string, error) {
	body, err := c.DoResponse(http.MethodGet, "/api/v4/user", "")
	if err != nil {
		return "", err
	}

	var user User
	err = json.Unmarshal(body, &user)
	if err != nil {
		return "", err
	}
	return user.Username, nil
}

func (c *Client) WaitingForApprove() ([]MR, error) {
	username, err := c.GetCurrentUsername()
	if err != nil {
		return nil, err
	}

	body, err := c.DoResponse(http.MethodGet, "/api/v4/merge_requests?state=opened&scope=all", "")
	if err != nil {
		return nil, err
	}

	var mrsResponse []MR
	err = json.Unmarshal(body, &mrsResponse)
	if err != nil {
		return nil, err
	}

	mrsResponse = lo.Filter(mrsResponse, func(mr MR, index int) bool {
		if mr.Author.Username == username {
			return false
		}
		found := lo.ContainsBy(mr.Assignees, func(user User) bool {
			return user.Username == username
		})
		if found {
			return true
		}
		found = lo.ContainsBy(mr.Reviewers, func(user User) bool {
			return user.Username == username
		})
		if found {
			return true
		}
		mentions := userMentionRegex.FindAllString(mr.Description, -1)
		return lo.ContainsBy(mentions, func(mention string) bool {
			return mention == "@"+username
		})
	})

	for i, mr := range mrsResponse {
		body, err := c.DoResponse(http.MethodGet, fmt.Sprintf("/api/v4/projects/%d/merge_requests/%d/notes?sort=asc&order_by=updated_at", mr.ProjectID, mr.IID), "")
		if err != nil {
			return nil, err
		}
		var notesResponse []Note
		err = json.Unmarshal(body, &notesResponse)
		if err != nil {
			return nil, err
		}
		mrsResponse[i].ApprovedBy = lo.FilterMap(notesResponse, func(note Note, _ int) (User, bool) {
			if note.Body != approveNoteBody {
				return User{}, false
			}
			return note.Author, true
		})
		mrsResponse[i].ApprovedByMe = lo.ContainsBy(mrsResponse[i].ApprovedBy, func(user User) bool {
			return user.Username == username
		})
	}

	return mrsResponse, nil
}

func (c *Client) Do(method, uri, body string) error {
	_, err := c.DoResponse(method, uri, body)
	return err
}

func (c *Client) DoResponse(method, uri, body string) ([]byte, error) {
	req, err := c.request(method, uri, body)
	if err != nil {
		return nil, err
	}

	return doRequest(req)
}

func (c *Client) request(method, uri, body string) (*http.Request, error) {
	url := fmt.Sprintf("https://%s%s", c.host, uri)

	buf := strings.NewReader(body)
	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+c.authToken)
	req.Header.Add("Content-Type", "application/json")
	return req, nil
}

func doRequest(req *http.Request) ([]byte, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode > http.StatusIMUsed {
		body := ""
		if resp.Body != nil {
			respBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("fail to read response body: %s", err.Error())
			}
			body = string(respBody)
		}
		return nil, fmt.Errorf("status code %d, body: %s", resp.StatusCode, body)
	}
	return ioutil.ReadAll(resp.Body)
}
