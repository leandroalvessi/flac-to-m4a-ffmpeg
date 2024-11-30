package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/rivo/tview"
)

var RenomearPorNumero = false // Variável global que define se deve renomear os arquivos por número

func main() {
	form()
}

func form() {
	app := tview.NewApplication()

	var form *tview.Form // Declaração da variável form

	tamanhoCampos := 60

	form = tview.NewForm().
		AddInputField("Diretório de Entrada", "C:\\Users\\leand\\Music", tamanhoCampos, nil, nil).
		AddInputField("Diretório de Saida", "C:\\Users\\leand\\Music", tamanhoCampos, nil, nil).
		AddInputField("Qualidade do áudio", "10", 10, nil, nil).
		AddCheckbox("Renomear Por Numero", false, func(checked bool) {
			RenomearPorNumero = checked // Atualiza a variável global com o estado da checkbox
		}).
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

	form.SetBorder(true).SetTitle(" Flac To M4a FFMpeg ").SetTitleAlign(tview.AlignCenter)
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

	var wg sync.WaitGroup

	// Obter o número de núcleos disponíveis na CPU
	numCPU := runtime.NumCPU()

	// Criar um canal com capacidade igual ao número de núcleos da CPU
	sem := make(chan struct{}, numCPU)

	// Lista de arquivos a serem processados
	var fileQueue []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".flac") {
			fileQueue = append(fileQueue, filepath.Join(inputDir, file.Name()))
		}
	}

	// Registrar o tempo de início
	startTime := time.Now()

	// Contador de arquivos processados
	totalFiles := len(fileQueue)

	// Exibir o progresso no terminal
	fmt.Printf("Iniciando conversão de %d arquivos FLAC...\n", totalFiles)

	// Processar os arquivos
	for idx, inputFile := range fileQueue {
		// Adicionar uma Goroutine para processar o arquivo
		wg.Add(1)

		// Adicionar um "token" ao semáforo
		sem <- struct{}{}

		// Iniciar a Goroutine para conversão
		go func(idx int, inputFile string) {
			defer wg.Done() // Decrementar o contador quando a Goroutine terminar

			// Gerar nome do arquivo de saída
			var outputFile string

			// Verificar se a variável RenomearPorNumero está ativada
			if RenomearPorNumero {
				// Formatar o número da música com dois dígitos
				numeroMusica := fmt.Sprintf("%02d", idx+1)
				outputFile = filepath.Join(outputDir, numeroMusica+" - "+strings.TrimSuffix(filepath.Base(inputFile), ".flac")+".m4a")
			} else {
				outputFile = filepath.Join(outputDir, strings.TrimSuffix(filepath.Base(inputFile), ".flac")+".m4a")
			}

			// Verificar se o arquivo de saída já existe e, caso exista, criar um nome único
			counter := 1
			for {
				if _, err := os.Stat(outputFile); err == nil {
					outputFile = filepath.Join(outputDir, fmt.Sprintf("%s (Copia %d).m4a", strings.TrimSuffix(filepath.Base(inputFile), ".flac"), counter))
					counter++
				} else {
					break
				}
			}

			// Executar a conversão com ffmpeg
			fmt.Printf("Convertendo: %s -> %s\n", inputFile, outputFile)

			cmd := exec.Command(
				"ffmpeg",
				"-i", inputFile, // Arquivo de entrada
				"-c:a", "aac", // Codec de áudio AAC
				"-q:a", Quality, // Qualidade do áudio
				"-map", "0", // Preserva todos os fluxos (áudio, capa, etc.)
				"-map_metadata", "0", // Preserva todos os metadados
				"-c:v", "mjpeg", // Especifica o formato da capa como JPEG (usado por padrão em MP3/MP4)
				"-disposition:v", "attached_pic", // Define o fluxo de imagem como "capa do álbum"
				"-avoid_negative_ts", "make_zero", // Corrige timestamps negativos
				outputFile, // Arquivo de saída
			)

			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			err := cmd.Run()
			if err != nil {
				fmt.Printf("Erro ao converter %s: %v\n", inputFile, err)
				<-sem // Liberar o "token" no semáforo em caso de erro
				return
			}

			fmt.Printf("Arquivo convertido com sucesso: %s\n", outputFile)

			// Liberar o "token" no semáforo quando a tarefa terminar
			<-sem
		}(idx, inputFile)
	}

	// Esperar até que todas as Goroutines terminem
	wg.Wait()

	// Registrar o tempo de término
	endTime := time.Now()

	// Calcular o tempo total gasto
	duration := endTime.Sub(startTime)

	// Exibir a mensagem modal com o tempo de processamento
	modal(fmt.Sprintf("Conversão concluída com sucesso!\nTempo total: %s", duration))
}

func modal(text string) {
	app := tview.NewApplication()
	modal := tview.NewModal().
		SetText(text).
		AddButtons([]string{"Sair"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Sair" {
				app.Stop()
			}
		})
	if err := app.SetRoot(modal, false).SetFocus(modal).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
