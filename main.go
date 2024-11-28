package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rivo/tview"
)

func main() {
	form()
}

func form() {
	app := tview.NewApplication()

	var form *tview.Form // Declaração da variável form

	tamanhoCampos := 60

	form = tview.NewForm().
		AddInputField("Diretório de Entrada", "C:\\Users\\leand\\Downloads\\Telegram Desktop", tamanhoCampos, nil, nil).
		AddInputField("Diretório de Saida", "C:\\Users\\leand\\Downloads\\Telegram Desktop", tamanhoCampos, nil, nil).
		AddInputField("Qualidade do áudio", "10", 20, nil, nil).
		AddButton("Converter", func() {
			app.Stop()
			inputFile := form.GetFormItemByLabel("Diretório de Entrada").(*tview.InputField).GetText()
			outputFile := form.GetFormItemByLabel("Diretório de Saida").(*tview.InputField).GetText()
			audioQuality := form.GetFormItemByLabel("Qualidade do áudio").(*tview.InputField).GetText()
			converter(inputFile, outputFile, audioQuality)
		}).
		AddButton("Quit", func() {
			app.Stop()
		})

	form.SetBorder(true).SetTitle(" Entre com as informações ").SetTitleAlign(tview.AlignCenter)
	if err := app.SetRoot(form, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func converter(inputDir, outputDir, Quality string) {
	files, err := os.ReadDir(inputDir)
	if err != nil {
		fmt.Printf("Erro ao abrir o diretório: %v\n", err)
		return
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".flac") {
			inputFile := filepath.Join(inputDir, file.Name())
			outputFile := filepath.Join(outputDir, strings.TrimSuffix(file.Name(), ".flac")+".m4a")

			fmt.Printf("Convertendo: %s -> %s\n", inputFile, outputFile)

			cmd := exec.Command(
				"ffmpeg",
				"-i", inputFile, // Arquivo de entrada
				"-vn",         // Ignorar vídeo
				"-c:a", "aac", // Codec de áudio AAC
				"-q:a", Quality, // Qualidade do áudio
				"-map_metadata", "-1", // Remove metadados
				outputFile, // Arquivo de saída
			)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			// Executar o comando FFmpeg
			err := cmd.Run()
			if err != nil {
				fmt.Printf("Erro ao converter %s: %v\n", inputFile, err)
				continue
			}

			fmt.Printf("Arquivo convertido com sucesso: %s\n", outputFile)
		}
	}

	fmt.Println("Conversão concluída para todos os arquivos FLAC!")
}
