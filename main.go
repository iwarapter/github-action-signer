package main

import (
	"context"
	"encoding/base64"
	"github.com/go-git/go-git/v5"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
	"io/ioutil"
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
	changes := &[]githubv4.FileAddition{}
	for name, status := range s {
		if status.Worktree == git.Modified {
			b, _ := ioutil.ReadFile(name)
			content := base64.StdEncoding.EncodeToString(b)
			*changes = append(*changes, githubv4.FileAddition{
				Path:     githubv4.String(name),
				Contents: githubv4.Base64String(content),
			})
		}
	}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	client := githubv4.NewClient(httpClient)

	var m struct {
		CreateCommitOnBranch struct {
			Commit struct {
				Url githubv4.ID
			}
		} `graphql:"createCommitOnBranch(input: $input)"`
	}
	input := githubv4.CreateCommitOnBranchInput{
		Branch: githubv4.CommittableBranch{
			RepositoryNameWithOwner: githubv4.NewString(githubv4.String(os.Getenv("GITHUB_REPOSITORY"))),
			BranchName:              githubv4.NewString(githubv4.String(os.Getenv("GITHUB_REF_NAME"))),
		},
		Message: githubv4.CommitMessage{Headline: "this is a test"},
		FileChanges: &githubv4.FileChanges{
			Additions: changes,
		},
		ExpectedHeadOid: githubv4.GitObjectID(rev.Hash().String()),
	}

	err = client.Mutate(context.Background(), &m, input, nil)
	if err != nil {
		log.Fatalf("unable to mutate: %s", err)
	}
	log.Printf("mutation complete: %s", m.CreateCommitOnBranch.Commit.Url)

}
