package main

import (
	"fmt"
	"os"

	"github.com/Madiyanke/gh-sentinel/internal/ai"
	"github.com/Madiyanke/gh-sentinel/internal/api"
	"github.com/Madiyanke/gh-sentinel/internal/ui"
	"github.com/charmbracelet/bubbles/list"
)

func main() {
	client, err := api.NewClient()
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🔍 Sentinel scanne : %s/%s\n", client.Owner, client.Repo)
	runs, err := client.GetFailedWorkflows()
	if err != nil {
		fmt.Printf("❌ Erreur API : %v\n", err)
		os.Exit(1)
	}

	if len(runs) == 0 {
		fmt.Println("✅ Aucun échec détecté. Votre CI est saine !")
		return
	}

	var items []list.Item
	for _, r := range runs {
		items = append(items, ui.Item{
			TitleStr: r.GetDisplayTitle(),
			DescStr:  fmt.Sprintf("Event: %s | ID: %d", r.GetEvent(), r.GetID()),
			ID:       r.GetID(),
		})
	}

	selected := ui.StartSelector(items)
	if selected == nil {
		fmt.Println("👋 Opération annulée.")
		return
	}

	fmt.Printf("🤖 Analyse de l'échec ID %d via Copilot...\n", selected.ID)
	logs, _ := client.GetJobLogs(selected.ID)
	
	suggestion, err := ai.SuggestFix(logs)
	if err != nil {
		fmt.Printf("❌ Erreur IA : %v\n", err)
		return
	}

	fmt.Println("\n" + suggestion)
}