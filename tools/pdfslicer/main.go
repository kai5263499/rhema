package main

import (
	"bufio"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	destinationPdf string
)

func main() {
	if len(os.Args) < 2 {
		logrus.Fatalf("usage: %s <pdfname>", os.Args[0])
	}
	logrus.SetFormatter(&logrus.TextFormatter{})

	logrus.SetLevel(logrus.DebugLevel)

	flag.StringVar(&destinationPdf, "dest", "./", "where to save split out pdfs")
	flag.Parse()

	pdfName := os.Args[1]

	logrus.Debugf("parsing %s", pdfName)

	sections, err := getSections(pdfName)
	if err != nil {
		logrus.WithError(err).Fatal("error getting sections")
	}

	logrus.Debugf("parsing %d sections", len(sections))

	for _, section := range sections {
		if err := splitPDF(pdfName, section); err != nil {
			logrus.WithError(err).Fatal("error getting sections")
		}
	}
}

func splitPDF(sourcePdf string, section *Section) error {
	destPdf := filepath.Join(destinationPdf, section.name+".pdf")

	if section.pageStart == 0 {
		section.pageStart++
	}

	cmd := exec.Command("pdftk", sourcePdf, "cat", strconv.Itoa(section.pageStart)+"-"+strconv.Itoa(section.pageEnd), "output", destPdf)
	logrus.Debugf("running pdftk with args %+v", cmd.Args)
	if err := cmd.Run(); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"pageStart": section.pageStart,
			"pageEnd":   section.pageEnd,
			"sourcePdf": sourcePdf,
			"destPdf":   destPdf,
		}).Error("error splitting pdf")
		return err
	}

	return nil
}

type Section struct {
	name      string
	pageStart int
	pageEnd   int
}

func getSections(sourcePdf string) ([]*Section, error) {
	logrus.Debugf("running: mutool show %s outline", sourcePdf)

	cmd := exec.Command("mutool", "show", sourcePdf, "outline")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	sections := []*Section{}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	pageRegex := regexp.MustCompile(`#([0-9]+),[0-9]+,[0-9]+`)
	nameRegex := regexp.MustCompile(`"([^"]*)"`)

	var currentSection *Section

	for scanner.Scan() {
		line := scanner.Text()

		logrus.Debugf("parsing line=%s", line)

		pageMatch := pageRegex.FindStringSubmatch(line)
		nameMatch := nameRegex.FindStringSubmatch(line)

		if len(pageMatch) > 0 && len(nameMatch) > 0 {
			currentSection = &Section{
				name: nameMatch[1],
			}

			if currentSection != nil {
				currentSection.pageStart, err = strconv.Atoi(pageMatch[1])
				if err != nil {
					logrus.WithError(err).Fatalf("error converting %s to int", pageMatch[1])
				}
				sections = append(sections, currentSection)

				if len(sections) > 2 {
					lastSectionPtr := len(sections) - 2
					sections[lastSectionPtr].pageEnd = currentSection.pageStart - 1
				}

				sections = append(sections, currentSection)
			}
		}
	}

	if currentSection != nil {
		// Assuming the last section goes until the end of the document.
		currentSection.pageEnd = -1
		sections = append(sections, currentSection)
	}

	return sections, nil
}
