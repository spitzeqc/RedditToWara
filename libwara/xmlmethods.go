package libwara

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// Creates an empty Nup
func InitNup() *Nup {
	ret := Nup{
		Version:     1,
		HasError:    0,
		RequestName: "topics",
		Expire:      "2100-01-01 10:00:00",
		Topics:      []Topic{},

		totalPosts: 0,
	}

	return &ret
}

// Topic Methods
// Adds an empty Topic with a specified name to the Nup
func (n *Nup) AddTopic(name string) error {
	if n == nil {
		return errors.New("no Nup provided")
	}
	if len(n.Topics) > 10 {
		return errors.New("too many Topics")
	}

	for i := 0; i < len(n.Topics); i++ {
		if n.Topics[i].Name == name {
			return errors.New("Topic with this name already exists")
		}
	}

	t := Topic{
		Icon:             defaultIcon,
		TitleId:          defaultTitleIds[len(n.Topics)],
		CommunityId:      defaultCommunityId,
		IsRecommended:    0,
		Name:             name,
		ParticipantCount: 0,
		Posts:            []Post{},
		EmpathyCount:     0,
		HasShopPage:      0,
		ModifiedAt:       time.Now().Format("2006-01-02 15:04:05"),
		Position:         2,
	}

	n.Topics = append(n.Topics, t)

	return nil
}

// Returns a Topic with a given name from the Nup
func (n *Nup) GetTopic(name string) (*Topic, error) {
	if n == nil {
		return nil, errors.New("no Nup provided")
	}
	for i, t := range n.Topics {
		if t.Name == name {
			return &n.Topics[i], nil
		}
	}

	return nil, errors.New("could not find Topic with name " + name)
}

// Removes a Topic with a given name from the Nup
func (n *Nup) RemoveTopic(name string) error {
	if n == nil {
		return errors.New("no Nup provided")
	}

	for i, t := range n.Topics {
		if t.Name == name {
			n.Topics = append(n.Topics[:i], n.Topics[i+1:]...)
			return nil
		}
	}

	return errors.New("could not find Topic with name " + name)
}

// Adds an empty post to a topic
func (n *Nup) AddPost(topicName string) (*Post, error) {
	t, err := n.GetTopic(topicName)
	if err != nil {
		return nil, err
	}
	if n.totalPosts >= 260 {
		return nil, errors.New("too many Posts")
	}
	t.Posts = append(t.Posts, Post{
		Body:        "Blank post",
		CommunityId: int(t.CommunityId),
		CountryId:   1,
		CreatedAt:   time.Now().Format("2006-01-02 15:04:05"),
		FeelingId:   FEELING_DEFAULT,
		LanguageId:  1,
		MiiData:     defaultMii,
		PlatformId:  1,
		ScreenName:  "Blank",
		PaintingUrl: "http://botu",
	})
	n.totalPosts++
	return &t.Posts[len(t.Posts)-1], nil
}

// Removes a post from a topic
func (n *Nup) RemovePost(topicName string, index uint) error {
	t, err := n.GetTopic(topicName)
	if err != nil {
		return err
	}

	if index >= MAX_POSTS {
		var msg strings.Builder
		msg.WriteString("index ")
		msg.WriteString(strconv.FormatUint(uint64(index), 10))
		msg.WriteString(" out of range for topic ")
		msg.WriteString(topicName)
		return errors.New(msg.String())
	}

	t.Posts = append(t.Posts[:index], t.Posts[index+1:]...)

	return nil
}

// Temporary, for testing only
func SetAllMiis(xmlPath string, miiString string) {
	xmlInBytes, err := os.ReadFile(xmlPath)
	if err != nil {
		log.Fatal(err)
	}

	x := &Nup{}
	err = xml.Unmarshal(xmlInBytes, x)
	if err != nil {
		log.Fatal(err)
	}

	for t := range x.Topics {
		for q := range x.Topics[t].Posts {
			x.Topics[t].Posts[q].MiiData = miiString
		}
	}

	xmlBytes, err := xml.MarshalIndent(x, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	for _, c := range replaceTable {
		xmlBytes = []byte(strings.Replace(string(xmlBytes[:]), c.Target, c.Replacement, -1))
	}

	outFile, err := os.Create("1stNUP.xml")
	if err != nil {
		log.Fatal(err)
	}
	_, err = outFile.Write([]byte(xml.Header))
	if err != nil {
		log.Fatal(err)
	}
	_, err = outFile.Write(xmlBytes)
	if err != nil {
		log.Fatal(err)
	}
}

// "Renders" a Nup into a string. If a path is provided, the output is written to the file
func (n *Nup) Render(outputName ...string) (string, error) {
	write := func(w io.Writer, b *[]byte) error {
		_, err := w.Write([]byte(xml.Header))
		if err != nil {
			return err
		}
		_, err = w.Write((*b)[:])
		if err != nil {
			return err
		}

		return nil
	}

	xmlBytes, err := xml.MarshalIndent(n, "", "  ")
	if err != nil {
		return "", err
	}

	for _, c := range replaceTable {
		xmlBytes = []byte(strings.Replace(string(xmlBytes[:]), c.Target, c.Replacement, -1))
	}

	if len(outputName) == 0 {
		out := &bytes.Buffer{}
		if err = write(out, &xmlBytes); err != nil {
			return "", err
		}
		return (*out).String(), nil
	}

	out, err := os.Create(outputName[0])
	if err != nil {
		return "", err
	}
	if err = write(out, &xmlBytes); err != nil {
		return "", err
	}

	return "", nil
}
