package add

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vanducng/jira-cli/pkg/jira"
)

type fakeCommentClient struct {
	uploaded   []string
	comment    string
	imageNames []string
	commentErr error
}

func (f *fakeCommentClient) AddIssueAttachment(_ string, path string) (*jira.IssueAttachment, error) {
	f.uploaded = append(f.uploaded, path)
	index := len(f.uploaded)
	return &jira.IssueAttachment{ID: fmt.Sprint(index), Filename: fmt.Sprintf("image-%d.png", index)}, nil
}

func (f *fakeCommentClient) AddIssueComment(_ string, comment string, _ bool, imageNames ...string) error {
	f.comment = comment
	f.imageNames = imageNames
	return f.commentErr
}

func TestNewCmdCommentAddImageFlag(t *testing.T) {
	flag := NewCmdCommentAdd().Flags().Lookup("image")
	require.NotNil(t, flag)
	assert.Equal(t, "i", flag.Shorthand)
	assert.Equal(t, "stringArray", flag.Value.Type())
}

func TestSubmitUploadsImagesBeforeComment(t *testing.T) {
	client := &fakeCommentClient{}
	command := addCmd{
		client: client,
		params: &addParams{
			issueKey: "TEST-1",
			body:     "Implementation flow",
			images:   []string{"before.png", "after.png"},
		},
	}

	err := command.submit()
	assert.NoError(t, err)
	assert.Equal(t, []string{"before.png", "after.png"}, client.uploaded)
	assert.Equal(t, "Implementation flow", client.comment)
	assert.Equal(t, []string{"image-1.png", "image-2.png"}, client.imageNames)
}

func TestSubmitReportsPartialWrite(t *testing.T) {
	client := &fakeCommentClient{commentErr: errors.New("comment failed")}
	command := addCmd{
		client: client,
		params: &addParams{
			issueKey: "TEST-1",
			images:   []string{"before.png", "after.png"},
		},
	}

	err := command.submit()
	assert.ErrorContains(t, err, "after uploading attachments 1, 2")
	assert.ErrorIs(t, err, client.commentErr)
}
