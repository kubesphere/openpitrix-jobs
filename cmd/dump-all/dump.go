// Copyright 2021 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"

	"openpitrix.io/openpitrix/pkg/client/internals3"
	"openpitrix.io/openpitrix/pkg/config"
	"openpitrix.io/openpitrix/pkg/db"
	"openpitrix.io/openpitrix/pkg/models"
	"openpitrix.io/openpitrix/pkg/pi"
)

func init() {
	_ = os.MkdirAll(filepath.Join(os.TempDir(), "op-dump"), os.ModeDir)
}

func getFilePath(filename string) string {
	return filepath.Join(os.TempDir(), "op-dump", filename)
}

func CtxWithDB(database string, ctx context.Context) context.Context {
	cfg := config.GetConf()
	cfg.Mysql.Database = database
	pi.SetGlobal(cfg)
	return db.NewContext(ctx, cfg.Mysql)
}

func errHandler(err error) {
	if err != nil {
		panic(err)
	}
}

func writeFileContent(filename string, content interface{}) {
	outputFile, err := os.Create(getFilePath(filename))
	errHandler(err)

	encoder := json.NewEncoder(outputFile)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(content)
	errHandler(err)

	log.Printf("Success to write [%d] rows to file [%s]", reflect.ValueOf(content).Len(), filename)
}

func listAttachmentFilenames(ctx context.Context, attachment models.Attachment) ([]string, error) {
	// with prefix
	var filenames []string
	output, err := internals3.S3.ListObjectsWithContext(ctx, &s3.ListObjectsInput{
		Bucket: internals3.Bucket,
		Prefix: aws.String(attachment.GetObjectPrefix()),
	})
	if err != nil {
		return nil, err
	}
	for _, o := range output.Contents {
		if o.Key != nil {
			filenames = append(filenames, attachment.RemoveObjectName(*o.Key))
		}
	}
	return filenames, nil
}

func dumpApps() {
	var rows []models.App
	ctx := CtxWithDB("app", context.Background())
	_, err := pi.Global().DB(ctx).Select("*").Where(db.Neq("status", "deleted")).From("app").Load(&rows)
	errHandler(err)
	writeFileContent("apps.json", rows)
}
func dumpAppVersions() {
	var rows []models.AppVersion
	ctx := CtxWithDB("app", context.Background())
	_, err := pi.Global().DB(ctx).Select("*").Where(db.Neq("status", "deleted")).From("app_version").Load(&rows)
	errHandler(err)
	writeFileContent("app_versions.json", rows)
}
func dumpCategories() {
	var rows []models.Category
	ctx := CtxWithDB("app", context.Background())
	_, err := pi.Global().DB(ctx).Select("*").From("category").Load(&rows)
	errHandler(err)
	writeFileContent("categories.json", rows)
}
func dumpCategoryResources() {
	var rows []models.CategoryResource
	ctx := CtxWithDB("app", context.Background())
	_, err := pi.Global().DB(ctx).Select("*").From("category_resource").Load(&rows)
	errHandler(err)
	writeFileContent("category_resources.json", rows)
}
func dumpClusters() {
	var rows []models.Cluster
	ctx := CtxWithDB("cluster", context.Background())
	_, err := pi.Global().DB(ctx).Select("*").Where(db.Neq("status", []string{"deleted", "ceased"})).From("cluster").Load(&rows)
	errHandler(err)
	writeFileContent("clusters.json", rows)
}
func dumpRepos() {
	var rows []models.Repo
	ctx := CtxWithDB("repo", context.Background())
	_, err := pi.Global().DB(ctx).Select("*").Where(db.Neq("status", "deleted")).From("repo").Load(&rows)
	errHandler(err)
	writeFileContent("repos.json", rows)
}
func dumpRepoLabels() {
	var rows []models.RepoLabel
	ctx := CtxWithDB("repo", context.Background())
	_, err := pi.Global().DB(ctx).Select("*").From("repo_label").Load(&rows)
	errHandler(err)
	writeFileContent("repo_labels.json", rows)
}
func dumpAttachments() {
	var rows []models.Attachment
	ctx := CtxWithDB("attachment", context.Background())
	_, err := pi.Global().DB(ctx).Select("*").From("attachment").Load(&rows)
	errHandler(err)
	writeFileContent("attachments.json", rows)

	for _, attachment := range rows {
		filenames, err := listAttachmentFilenames(ctx, attachment)
		errHandler(err)

		for _, filename := range filenames {
			content, err := internals3.S3.GetObjectWithContext(ctx, &s3.GetObjectInput{
				Bucket: internals3.Bucket,
				Key:    aws.String(attachment.GetObjectName(filename)),
			})
			errHandler(err)

			var path = filepath.Join(os.TempDir(), "op-dump", attachment.AttachmentId, filename)
			err = os.MkdirAll(filepath.Dir(path), os.ModeDir)
			errHandler(err)

			file, err := os.Create(path)
			errHandler(err)
			_, err = io.Copy(file, content.Body)
			errHandler(err)
		}
	}
}

func main() {

	dumpApps()
	dumpAppVersions()
	dumpCategories()
	dumpCategoryResources()
	dumpClusters()
	dumpRepos()
	dumpRepoLabels()
	dumpAttachments()
}
