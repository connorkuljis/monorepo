package gotenberg

func GeneratePDF(filename string) {
	// 	// 4.  Opening the rendered cover letter to send to gotenburg to format as pdf
	// 	// NOTE: this is a hack as goteberg does not allow a converting html bytes directly to pdf as it must be a form file in the header.
	// 	filename = filepath.Join("out", filename, "index.html")
	// 	file, err := os.Open(filename)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	defer file.Close()

	// 	body := &bytes.Buffer{}
	// 	writer := multipart.NewWriter(body)

	// 	// 4.1 Allocating and copying the form fil
	// 	part, err := writer.CreateFormFile("files", filepath.Base(filename))
	// 	if err != nil {
	// 		return err
	// 	}

	// 	_, err = io.Copy(part, file)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	err = writer.Close()
	// 	if err != nil {
	// 		return err
	// 	}

	// 	// 4.2 Posting the form file to gotenberg
	// 	url := h.GotenbergServiceURL + "/forms/chromium/convert/html"
	// 	req, err := http.NewRequest("POST", url, body)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 	client := &http.Client{}
	// 	resp2, err := client.Do(req)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	defer resp2.Body.Close()
	// 	if resp2.StatusCode != http.StatusOK {
	// 		return err
	// 	}

	// 	// 4.3 Respond with pdf
	// 	c.Response().Header().Set("Content-Type", "application/pdf")
	// 	c.Response().Header().Set("Content-Disposition", "attachment; filename=converted.pdf")

	// 	_, err = io.Copy(c.Response(), resp2.Body)
	// 	if err != nil {
	// 		return err
	// 	}
}
