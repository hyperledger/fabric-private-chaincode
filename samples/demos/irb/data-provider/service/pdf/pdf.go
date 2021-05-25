package pdf

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func ParseQuestionForm(pdfContent []byte) (*PatientInformation, error) {
	// create working dir
	workingDir, err := ioutil.TempDir("", "parseConsentPdf")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(workingDir)

	input := filepath.Join(workingDir, "input.pdf")
	output := filepath.Join(workingDir, "output.txt")

	// store pdfContent as input.pdf
	err = ioutil.WriteFile(input, pdfContent, 0644)
	if err != nil {
		return nil, err
	}

	// translate
	pdfToText(input, output)

	// read txt output
	content, err := ioutil.ReadFile(output)
	if err != nil {
		return nil, err
	}

	// parse
	answers, err := parseAnswers(string(content))
	if err != nil {
		return nil, err
	}

	return answers, nil
}

func pdfToText(input, output string) {

	// apt install poppler-utils
	//pdftotext -layout Patient_Consent_Form_1.pdf  output.txt

	command := "pdftotext"
	flags := "-layout"

	// trigger pdftotext
	fmt.Printf("Executing %s %s %s %s\n", command, flags, input, output)
	cmd := exec.Command(command, flags, input, output)
	_, err := cmd.Output()
	if err != nil {
		panic(err)
	}
}

type PatientInformation struct {
	StudyID  string
	UUID     string
	Name     string
	Birthday string
	Gender   string
	Answers  *PatientAnswers
}

type PatientAnswers struct {
	Answers []float32
}

func (a *PatientAnswers) ToString() string {
	var sb strings.Builder
	for _, ans := range a.Answers {
		sb.WriteString(fmt.Sprintf("%g, ", ans))
	}
	return strings.Trim(sb.String(), ", ")
}

func parseAnswers(content string) (*PatientInformation, error) {

	// TODO find study ID

	// find patient Id
	uuid, err := search(content, "Patient Id:")
	if err != nil {
		return nil, err
	}

	// find name
	name, err := search(content, "Name:")
	if err != nil {
		return nil, err
	}

	// find birthday
	birthdate, err := search(content, "Birthdate:")
	if err != nil {
		return nil, err
	}

	// find gender
	gender, err := search(content, "Gender:")
	if err != nil {
		return nil, err
	}

	questions := []string{
		"Temperature:",
		"Occurrence of nausea:",
		"Lumbar pain:",
		"Urine pushing (continuous need for urination):",
		"Micturition pains:",
		"Burning of urethra, itch, swelling of urethra outlet:",
		"Decision: Inflammation of urinary bladder:",
		"Decision: Nephritis of renal pelvis origin:",
	}
	answers, err := findAnswers(content, questions)
	if err != nil {
		return nil, err
	}

	return &PatientInformation{
		UUID:     uuid,
		Name:     name,
		Birthday: birthdate,
		Gender:   gender,
		Answers:  answers,
	}, nil

}

func findAnswers(content string, questions []string) (*PatientAnswers, error) {
	pa := &PatientAnswers{}
	for _, s := range questions {
		q, err := search(content, s)
		if err != nil {
			return nil, err
		}
		pa.Answers = append(pa.Answers, toFloat32(q))
	}
	return pa, nil
}

func toFloat32(answer string) float32 {
	switch answer {
	case "no":
		return float32(0)
	case "yes":
		return float32(1)
	default:
		// should not happen ... too lazy to write a proper error here
		answerFloat64, _ := strconv.ParseFloat(answer, 32)
		return float32(answerFloat64)
	}
}

func search(content, search string) (string, error) {
	for _, line := range strings.Split(content, "\n") {
		if strings.Index(line, search) > -1 {
			splits := strings.Split(line, search)
			return strings.TrimSpace(splits[1]), nil
		}
	}
	return "", fmt.Errorf("field not found: %s", search)
}
