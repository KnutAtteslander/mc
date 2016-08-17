/*
 * Minio Client (C) 2016 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mc

import (
	"encoding/json"

	"github.com/fatih/color"
	"github.com/minio/cli"
	"github.com/minio/mc/pkg/console"
	"github.com/minio/minio/pkg/probe"
)

var (
	eventsListFlags = []cli.Flag{}
)

var eventsListCmd = cli.Command{
	Name:   "list",
	Usage:  "List bucket notifications.",
	Action: mainEventsList,
	Flags:  append(eventsListFlags, globalFlags...),
	CustomHelpTemplate: `NAME:
   mc events {{.Name}} - {{.Usage}}

USAGE:
   mc events {{.Name}} ALIAS/BUCKET ARN [FLAGS]

FLAGS:
  {{range .Flags}}{{.}}
  {{end}}
EXAMPLES:
   1. List notification configurations associated to a specific arn
     $ mc events {{.Name}} myminio/mybucket arn:aws:sqs:us-west-2:444455556666:your-queue 
   2. List all notification configurations
     $ mc events {{.Name}} s3/mybucket
`,
}

// checkEventsListSyntax - validate all the passed arguments
func checkEventsListSyntax(ctx *cli.Context) {
	if len(ctx.Args()) != 2 && len(ctx.Args()) != 1 {
		cli.ShowCommandHelpAndExit(ctx, "list", 1) // last argument is exit code
	}
}

// eventsListMessage container
type eventsListMessage struct {
	Status string   `json:"status"`
	ID     string   `json:"id"`
	Events []string `json:"events"`
	Prefix string   `json:"prefix"`
	Suffix string   `json:"suffix"`
	Arn    string   `json:"arn"`
}

func (u eventsListMessage) JSON() string {
	u.Status = "success"
	eventsListMessageJSONBytes, e := json.Marshal(u)
	fatalIf(probe.NewError(e), "Unable to marshal into JSON.")
	return string(eventsListMessageJSONBytes)
}

func (u eventsListMessage) String() string {
	msg := console.Colorize("Events", u.ID+"\t"+u.Arn+"\t")
	for i, event := range u.Events {
		msg += console.Colorize("Events", event)
		if i != len(u.Events)-1 {
			msg += ","
		}
	}
	if u.Prefix != "" {
		msg += "\tprefix:" + u.Prefix
	}
	if u.Suffix != "" {
		msg += "\tsuffix:" + u.Suffix
	}
	return msg
}

func mainEventsList(ctx *cli.Context) {
	console.SetColor("Events", color.New(color.FgGreen, color.Bold))

	setGlobalsFromContext(ctx)
	checkEventsListSyntax(ctx)

	args := ctx.Args()
	path := args[0]
	arn := ""
	if len(args) > 1 {
		arn = args[1]
	}

	client, err := newClient(path)
	if err != nil {
		fatalIf(err.Trace(), "Cannot parse the provided url.")
	}

	s3Client, ok := client.(*s3Client)
	if !ok {
		fatalIf(errDummy().Trace(), "The provided url doesn't point to a S3 server.")
	}

	configs, err := s3Client.ListNotificationConfigs(arn)
	fatalIf(err, "Cannot enable notification on the specified bucket.")

	for _, config := range configs {
		printMsg(eventsListMessage{Events: config.Events,
			Prefix: config.Prefix,
			Suffix: config.Suffix,
			Arn:    config.Arn,
			ID:     config.ID})
	}
}
