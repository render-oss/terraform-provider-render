package common

import (
	"terraform-provider-render/internal/client"
)

type SecretFileModel struct {
	Content string `tfsdk:"content"`
}

func SecretFilesToClient(sfs map[string]SecretFileModel) []client.SecretFileInput {
	if len(sfs) == 0 {
		return nil
	}

	var res []client.SecretFileInput
	for k, v := range sfs {
		res = append(res, client.SecretFileInput{
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
