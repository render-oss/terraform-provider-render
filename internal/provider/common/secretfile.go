package common

import (
	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/client/envvar"
)

type SecretFileModel struct {
	Content string `tfsdk:"content"`
}

func SecretFilesToClient(sfs map[string]SecretFileModel) []envvar.SecretFileInput {
	if len(sfs) == 0 {
		return nil
	}

	var res []envvar.SecretFileInput
	for k, v := range sfs {
		res = append(res, envvar.SecretFileInput{
			Name:    k,
			Content: v.Content,
		})
	}

	return res
}

func SecretFilesFromClientCursors(sfs *[]client.SecretFileWithCursor) map[string]SecretFileModel {
	res := map[string]SecretFileModel{}

	if sfs == nil || len(*sfs) == 0 {
		return nil
	}

	for _, sf := range *sfs {
		res[sf.SecretFile.Name] = SecretFileModel{Content: sf.SecretFile.Content}
	}

	return res
}

func SecretFilesFromClient(sfs *[]client.SecretFile) map[string]SecretFileModel {
	res := map[string]SecretFileModel{}

	if sfs == nil || len(*sfs) == 0 {
		return nil
	}

	for _, sf := range *sfs {
		res[sf.Name] = SecretFileModel{Content: sf.Content}
	}

	return res
}
