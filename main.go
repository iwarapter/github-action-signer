package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
	"log"
	"os"
)

func main() {
	r, err := git.PlainOpen(".")
	if err != nil {
		log.Fatalf("unable to open repository: %s", err)
	}

	w, err := r.Worktree()
	if err != nil {
		log.Fatalf("unable to open repository: %s", err)
	}
	rev, err := r.Head()
	if err != nil {
		log.Fatalf("unable to find HEAD revision: %s", err)
	}
	s, err := w.Status()
	if err != nil {
		log.Fatalf("unable to open repository: %s", err)
	}
	log.Printf("status: %s", s.String())

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	client := githubv4.NewClient(httpClient)

	branch := "main"

	var m struct {
		AddReaction struct {
			Reaction struct {
				Content githubv4.ReactionContent
			}
			Subject struct {
				ID githubv4.ID
			}
		} `graphql:"addReaction(input: $input)"`
	}
	input := githubv4.CreateCommitOnBranchInput{
		Branch: githubv4.CommittableBranch{
			RepositoryNameWithOwner: githubv4.NewString("main"),
			BranchName:              githubv4.NewString(githubv4.String(branch)),
		},
		Message: githubv4.CommitMessage{Headline: "this is a test"},
		FileChanges: &githubv4.FileChanges{
			Additions: &[]githubv4.FileAddition{
				{
					Path:     "example.txt",
					Contents: githubv4.Base64String(base64.RawStdEncoding.EncodeToString([]byte("foo"))),
				},
			},
		},
		ExpectedHeadOid: githubv4.GitObjectID(rev.String()),
	}
	//input := githubv4.AddReactionInput{
	//	SubjectID: "targetIssue.ID", // ID of the target issue from a previous query.
	//	Content:   githubv4.ReactionContentHooray,
	//}

	err = client.Mutate(context.Background(), &m, input, nil)
	if err != nil {
		// Handle error.
	}
	fmt.Printf("Added a %v reaction to subject with ID %#v!\n", m.AddReaction.Reaction.Content, m.AddReaction.Subject.ID)

	// Output:
	// Added a HOORAY reaction to subject with ID "MDU6SXNzdWUyMTc5NTQ0OTc="!
}
