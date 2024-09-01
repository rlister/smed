package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"os"
	"os/exec"
	"sort"
)

var usage = "Usage: smed SECRETNAME [-lcv]"

var client *secretsmanager.Client

// create aws client
func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}
	client = secretsmanager.NewFromConfig(cfg)
}

func main() {
	create := flag.Bool("c", false, "view secrets")
	list := flag.Bool("l", false, "list secrets")
	view := flag.Bool("v", false, "view secrets")

	flag.Parse()
	args := flag.Args()

	switch {
	case *create:
		createSecrets(args)
	case *list:
		listSecrets(args)
	case *view:
		viewSecrets(args)
	case len(args) == 1:
		editSecret(args[0])
	default:
		fmt.Println(usage)
	}
}

// print sorted list of secrets (optionally matching strings)
func listSecrets(args []string) {
	filters := []types.Filter{}

	if len(args) > 0 {
		filters = append(filters, types.Filter{Key: "all", Values: args})
	}

	paginator := secretsmanager.NewListSecretsPaginator(client, &secretsmanager.ListSecretsInput{Filters: filters})
	secrets := []string{}

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(context.TODO())
		if err != nil {
			panic(err)
		}

		for _, s := range output.SecretList {
			secrets = append(secrets, *s.Name)
		}
	}

	sort.Strings(secrets)
	for _, s := range secrets {
		fmt.Println(s)
	}
}

// lookup external editor
func editor() string {
	ed, ok := os.LookupEnv("EDITOR")
	if !ok {
		ed = "vim"
	}
	return ed
}

// invoke external editor on tmpfile with value
func edit(data string) string {
	f, err := os.CreateTemp("", "smed-*.json")
	if err != nil {
		panic(err)
	}

	defer f.Close()
	defer os.Remove(f.Name())

	if _, err := f.Write([]byte(data)); err != nil {
		panic(err)
	}

	cmd := exec.Command(editor(), f.Name())
	if err = cmd.Run(); err != nil {
		panic(err)
	}

	// read back edited content
	edited, err := os.ReadFile(f.Name())
	if err != nil {
		panic(err)
	}

	return string(edited)
}

// create a new secret using editor
func createSecrets(names []string) {
	for _, name := range names {
		data := edit("{}")
		output, err := client.CreateSecret(context.TODO(), &secretsmanager.CreateSecretInput{Name: &name, SecretString: &data})
		if err == nil {
			fmt.Println(*output.ARN)
		} else {
			fmt.Println(err)
		}
	}
}

// edit existing secret in editor
func editSecret(name string) {
	val, err := client.GetSecretValue(context.Background(), &secretsmanager.GetSecretValueInput{SecretId: &name})
	if err != nil {
		fmt.Println(err)
		return
	}

	data := edit(*val.SecretString)
	output, err := client.PutSecretValue(context.TODO(), &secretsmanager.PutSecretValueInput{SecretId: &name, SecretString: &data})

	if err == nil {
		fmt.Println(*output.VersionId)
	} else {
		fmt.Println(err)
	}
}

// write secret value to stdout as formatted json
func viewSecrets(names []string) {
	for _, name := range names {
		output, err := client.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{
			SecretId: &name,
		})
		if err == nil {
			fmt.Println(formatJson(output.SecretString))
		} else {
			fmt.Println(err)
		}
	}
}

func formatJson(data *string) string {
	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(*data), "", "  "); err != nil {
		fmt.Println(err)
	}
	return buf.String()
}
