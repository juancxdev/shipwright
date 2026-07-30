package cmd

import "fmt"

func printDiscoveryChatGuidance() {
	fmt.Println("Discovery iniciado. Turno activo: product-owner.")
	fmt.Println()
	fmt.Println("El Product Owner NO debe inventar ni pedirte que llenes documentos a mano.")
	fmt.Println("Debe preguntarte en el chat lo que falte entender y recién después generar artefactos.")
	fmt.Println()
	fmt.Println("Abre OpenCode en este proyecto y ejecuta:")
	fmt.Println("  /shipwright-active-agent")
	fmt.Println()
	fmt.Println("O mandale este prompt al agente product-owner:")
	fmt.Println()
	fmt.Println("  Actúa como product-owner de Shipwright.")
	fmt.Println("  Lee .harness/artifacts/product/discovery.md y la petición inicial.")
	fmt.Println("  Hazme 3-7 preguntas de discovery en el chat antes de escribir contexto/scope.")
	fmt.Println("  Pregunta sobre usuarios, reglas de negocio, límites del MVP, flujo de facturas, estados y criterios de éxito.")
	fmt.Println("  No elijas arquitectura ni implementes código.")
	fmt.Println("  Registra las preguntas en .harness/artifacts/product/open-questions.md y espera mis respuestas si hay dudas críticas.")
	fmt.Println()
	fmt.Println("Cuando el PO ya tenga respuestas suficientes, debe generar:")
	fmt.Println("  .harness/artifacts/product/context.md")
	fmt.Println("  .harness/artifacts/product/assumptions.md")
	fmt.Println("  .harness/artifacts/product/open-questions.md")
	fmt.Println("  .harness/artifacts/product/scope.md")
	fmt.Println()
	fmt.Println("Después ejecuta: shipwright next")
}

func discoveryNextAction(missing []string) string {
	return fmt.Sprintf(`Blocked — Product Owner discovery round required.

Missing artifacts:
  %s

Do NOT start with shipwright scaffold unless you only want placeholders.
Recommended flow:
  1. Open OpenCode in this project.
  2. Run: /shipwright-active-agent
  3. Let product-owner ask discovery questions in chat.
  4. Answer the questions.
  5. product-owner writes .harness/artifacts/product/context.md, .harness/artifacts/product/assumptions.md, .harness/artifacts/product/open-questions.md and .harness/artifacts/product/scope.md.
  6. Run: shipwright next`, joinIndented(missing))
}

func joinIndented(items []string) string {
	if len(items) == 0 {
		return "(none)"
	}
	out := ""
	for i, item := range items {
		if i > 0 {
			out += "\n  "
		}
		out += item
	}
	return out
}
