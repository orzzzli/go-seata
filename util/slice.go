package util

func FilterSliceEmptyEle(input []string) []string {
	var output []string
	for _,v := range input {
		if v == "" || v == " " {
			continue
		}
		output = append(output,v)
	}
	return output
}
