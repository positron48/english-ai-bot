package main

type inputForm struct {
	Mood        string `json:"mood"`
	Tense       string `json:"tense"`
	Person      string `json:"person"`
	Number      string `json:"number"`
	Form        string `json:"form"`
	IsIrregular bool   `json:"is_irregular"`
}

type inputLemma struct {
	Lemma string      `json:"lemma"`
	Forms []inputForm `json:"forms"`
}
