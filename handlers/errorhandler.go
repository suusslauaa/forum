package handlers

import (
	"forum/templates"
	"html/template"
	"net/http"
)

type Errors struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func ErrorHandler(w http.ResponseWriter, message string, code int) {
	// Логирование ошибки для отладки (если присутствует)
	w.WriteHeader(code)

	// Загружаем шаблон ошибки
	tmpl, tmplErr := template.ParseFS(templates.Files, "error.html")
	if tmplErr != nil {
		// Если шаблон не загрузился, выводим простой текст ошибки
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Данные для передачи в шаблон
	errorData := Errors{
		Message: message,
		Code:    code,
	}

	// Рендеринг шаблона ошибки
	if err := tmpl.Execute(w, errorData); err != nil {
		http.Error(w, "Error rendering error page", http.StatusInternalServerError)
	}
}
