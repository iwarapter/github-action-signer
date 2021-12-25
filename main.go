package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
	"io/ioutil"
	"os"
)

func main() {
	r, err := git.PlainOpen(".")
	if err != nil {
		fmt.Printf("unable to open repository: %s", err)
		os.Exit(1)
	}

	w, err := r.Worktree()
	if err != nil {
		fmt.Printf("unable to open repository: %s", err)
		os.Exit(1)
	}
	rev, err := r.Head()
	if err != nil {
		fmt.Printf("unable to find HEAD revision: %s", err)
		os.Exit(1)
	}
	s, err := w.Status()
	if err != nil {
		fmt.Printf("unable to open repository: %s", err)
		os.Exit(1)
	}
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
	if len(*changes) == 0 {
		fmt.Printf("no changes, exiting")
		os.Exit(0)
	}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	client := githubv4.NewClient(httpClient)

	var branchName string
	if val, ok := os.LookupEnv("GITHUB_HEAD_REF"); ok {
		branchName = val
	} else {
		branchName = os.Getenv("GITHUB_REF_NAME")
	}

	message := "updated with commit signer"
	args := os.Args
	if len(args) > 1 {
		message = args[1]
	}

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
			BranchName:              githubv4.NewString(githubv4.String(branchName)),
		},
		Message: githubv4.CommitMessage{Headline: githubv4.String(message)},
		FileChanges: &githubv4.FileChanges{
			Additions: changes,
		},
		ExpectedHeadOid: githubv4.GitObjectID(rev.Hash().String()),
	}

	err = client.Mutate(context.Background(), &m, input, nil)
	if err != nil {
		fmt.Printf("unable to mutate: %s", err)
		os.Exit(1)
	}
	fmt.Printf("mutation complete: %s", m.CreateCommitOnBranch.Commit.Url)

}
