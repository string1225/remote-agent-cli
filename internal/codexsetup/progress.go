package codexsetup

type Progress struct {
	Stage      string
	Downloaded int64
	Total      int64
}

type downloadProgress struct {
	report     func(Progress)
	downloaded int64
	total      int64
}

func (p *downloadProgress) stage(stage string) {
	if p.report != nil {
		p.report(Progress{Stage: stage, Downloaded: p.downloaded, Total: p.total})
	}
}

func (p *downloadProgress) Write(data []byte) (int, error) {
	p.downloaded += int64(len(data))
	p.stage("download")
	return len(data), nil
}
