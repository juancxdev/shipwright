package cmd

import "fmt"

func printDiscoveryChatGuidance() {
	fmt.Println("Discovery iniciado. Turno activo: product-owner.")
	fmt.Println()
	fmt.Println("El Product Owner NO debe inventar ni pedirte que llenes documentos a mano.")
	fmt.Println("Debe preguntarte en el chat lo que falte entender y recién después generar artefactos.")
	fmt.Println()
	fmt.Println("Abrí OpenCode en este proyecto y ejecutá:")
	fmt.Println("  /shipwright-active-agent")
	fmt.Println()
	fmt.Println("O mandale este prompt al agente product-owner:")
	fmt.Println()
	fmt.Println("  Actuá como product-owner de Shipwright.")
	fmt.Println("  Leé product/discovery.md y la petición inicial.")
	fmt.Println("  Haceme 3-7 preguntas de discovery en el chat antes de escribir contexto/scope.")
	fmt.Println("  Preguntá sobre usuarios, reglas de negocio, límites del MVP, flujo de facturas, estados y criterios de éxito.")
	fmt.Println("  No elijas arquitectura ni implementes código.")
	fmt.Println("  Registrá las preguntas en product/open-questions.md y esperá mis respuestas si hay dudas críticas.")
	fmt.Println()
	fmt.Println("Cuando el PO ya tenga respuestas suficientes, debe generar:")
	fmt.Println("  product/context.md")
	fmt.Println("  product/assumptions.md")
	fmt.Println("  product/open-questions.md")
	fmt.Println("  product/scope.md")
	fmt.Println()
	fmt.Println("Después ejecutá: shipwright next")
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
  5. product-owner writes product/context.md, product/assumptions.md, product/open-questions.md and product/scope.md.
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
