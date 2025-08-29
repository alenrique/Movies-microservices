// Local: api-gateway/main.go

package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	httpSwagger "github.com/swaggo/http-swagger" // IMPORT DO SWAGGER

	_ "github.com/alenrique/Movies-microservices/api-gateway/docs"

	pb "github.com/alenrique/Movies-microservices/proto" // Importamos nosso pacote proto
)

// handler é uma struct que vai guardar nosso cliente gRPC.
// Usar uma struct para os handlers é uma boa prática para injetar dependências,
// como o cliente gRPC, de forma organizada.
type handler struct {
	client pb.MovieServiceClient
}

// MovieSwagger é uma struct apenas para documentação Swagger.
// Copie os campos do seu pb.Movie aqui.
type MovieSwagger struct {
	Id       string `json:"id"`
	Title    string `json:"title"`
	Director string `json:"director"`
	Year     int32  `json:"year"`
}

// CreateMovieRequestSwagger é uma struct apenas para documentação Swagger.
// Copie os campos do seu pb.CreateMovieRequest aqui.
type CreateMovieRequestSwagger struct {
	Title    string `json:"title"`
	Director string `json:"director"`
	Year     int32  `json:"year"`
}

// @title           API de Gerenciamento de Filmes
// @version         1.0
// @description     API REST para um sistema de microsserviços que gerencia filmes.
// @termsOfService  http://swagger.io/terms/

// @contact.name   Henrique Alencar
// @contact.url    https://github.com/alenrique
// @contact.email  henriquealencardev@gmail.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /
func main() {
	// --- Conexão gRPC (sem alterações) ---
	log.Println("Iniciando cliente gRPC para o Movie Service...")
	conn, err := grpc.NewClient("movies-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Não foi possível conectar ao servidor gRPC: %v", err)
	}
	defer conn.Close()
	client := pb.NewMovieServiceClient(conn)
	h := handler{client: client}

	// --- Configuração do Servidor HTTP (sem alterações) ---
	router := mux.NewRouter()

	// --- CORREÇÃO 2: Adicionando a rota do Swagger ---
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	router.HandleFunc("/movies", h.listMovies).Methods(http.MethodGet)
	router.HandleFunc("/movies", h.createMovie).Methods(http.MethodPost)
	router.HandleFunc("/movies/{id}", h.getMovie).Methods(http.MethodGet)
	router.HandleFunc("/movies/{id}", h.deleteMovie).Methods(http.MethodDelete)

	// --- NOVO: Lógica de Desligamento Gracioso (Graceful Shutdown) ---
	server := &http.Server{Addr: ":8080", Handler: router}

	// Canal para escutar por erros do servidor
	errChan := make(chan error, 1)

	// Inicia o servidor HTTP em uma goroutine separada
	go func() {
		log.Println("Servidor HTTP do API Gateway escutando na porta 8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Canal para escutar por sinais de interrupção do sistema (Ctrl+C)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Bloqueia a execução aqui até que um erro ou sinal seja recebido
	select {
	case err := <-errChan:
		log.Fatalf("Erro fatal no servidor HTTP: %v", err)
	case s := <-signalChan:
		log.Printf("Sinal '%v' recebido, iniciando desligamento gracioso...", s)

		// Cria um contexto com tempo limite para o desligamento
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		// Tenta desligar o servidor de forma ordenada
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("Erro durante o desligamento gracioso: %v", err)
		}
	}
}

// @Summary      Lista todos os filmes
// @Description  Retorna uma lista com todos os filmes cadastrados no banco de dados.
// @Tags         Filmes
// @Accept       json
// @Produce      json
// @Success      200  {array}   MovieSwagger "Lista de filmes"
// @Failure      500  {object}  object{error=string} "Erro interno no servidor"
// @Router       /movies [get]
func (h *handler) listMovies(w http.ResponseWriter, r *http.Request) {
	log.Println("Requisição recebida: GET /movies")

	// 1. Chamar o serviço gRPC
	// Usamos o contexto da requisição HTTP (r.Context()), que é uma boa prática
	// para propagar timeouts ou cancelamentos.
	res, err := h.client.ListMovies(r.Context(), &pb.ListMoviesRequest{})
	if err != nil {
		// Se a chamada gRPC falhar, retornamos um erro 500 (Internal Server Error).
		log.Printf("Erro ao chamar ListMovies via gRPC: %v", err)
		http.Error(w, "Erro interno ao buscar filmes", http.StatusInternalServerError)
		return
	}

	// 2. Escrever a resposta como JSON
	// Definimos o cabeçalho para indicar que a resposta é do tipo JSON.
	w.Header().Set("Content-Type", "application/json")

	// Codificamos a resposta (res.GetMovies()) diretamente no corpo da resposta HTTP (w).
	// As structs geradas pelo .proto já vêm com as tags `json:"..."`, então a conversão é automática.
	err = json.NewEncoder(w).Encode(res.GetMovies())
	if err != nil {
		log.Printf("Erro ao codificar resposta JSON: %v", err)
		http.Error(w, "Erro interno ao preparar resposta", http.StatusInternalServerError)
	}
}

// @Summary      Cria um novo filme
// @Description  Adiciona um novo filme à coleção a partir dos dados enviados no corpo da requisição.
// @Tags         Filmes
// @Accept       json
// @Produce      json
// @Param        movie  body      CreateMovieRequestSwagger  true  "Dados do Filme para Criar"
// @Success      201    {object}  MovieSwagger "Filme criado com sucesso"
// @Failure      400    {object}  object{error=string} "Requisição inválida"
// @Failure      500    {object}  object{error=string} "Erro interno no servidor"
// @Router       /movies [post]
func (h *handler) createMovie(w http.ResponseWriter, r *http.Request) {
	log.Println("Requisição recebida: POST /movies")

	// 1. Decodificar o JSON da requisição
	// Criamos uma variável do tipo que esperamos receber do cliente.
	var req pb.CreateMovieRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		// Se houver um erro na decodificação (ex: JSON mal formatado),
		// retornamos um erro 400 (Bad Request).
		http.Error(w, "Corpo da requisição inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 2. Chamar o serviço gRPC
	// Passamos o objeto 'req' que acabamos de preencher com os dados do JSON.
	res, err := h.client.CreateMovie(r.Context(), &req)
	if err != nil {
		log.Printf("Erro ao chamar CreateMovie via gRPC: %v", err)
		http.Error(w, "Erro interno ao criar o filme", http.StatusInternalServerError)
		return
	}

	// 3. Escrever a resposta de sucesso
	w.Header().Set("Content-Type", "application/json")
	// Para uma criação bem-sucedida, o status HTTP correto é 201 Created.
	w.WriteHeader(http.StatusCreated)
	// Codificamos o filme criado (que o gRPC nos retornou) como JSON na resposta.
	json.NewEncoder(w).Encode(res)
}

// @Summary      Busca um filme por ID
// @Description  Retorna os detalhes de um filme específico com base no seu ID.
// @Tags         Filmes
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "ID do Filme"
// @Success      200  {object}  MovieSwagger "Filme encontrado"
// @Failure      404  {object}  object{error=string} "Filme não encontrado"
// @Failure      500  {object}  object{error=string} "Erro interno no servidor"
// @Router       /movies/{id} [get]
func (h *handler) getMovie(w http.ResponseWriter, r *http.Request) {
	log.Println("Requisição recebida: GET /movies/{id}")

	// 1. Extrair o ID da URL
	// O gorilla/mux nos permite pegar variáveis da URL, como o {id}.
	vars := mux.Vars(r)
	id := vars["id"]

	// 2. Chamar o serviço gRPC
	req := &pb.GetMovieRequest{Id: id}
	res, err := h.client.GetMovie(r.Context(), req)
	if err != nil {
		// 3. Traduzir o erro do gRPC para um erro HTTP
		// Verificamos se o erro retornado pelo gRPC é do tipo 'NotFound'.
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			// Se for, retornamos um erro HTTP 404 Not Found.
			http.Error(w, st.Message(), http.StatusNotFound)
		} else {
			// Para qualquer outro erro, retornamos um 500 genérico.
			log.Printf("Erro ao chamar GetMovie via gRPC: %v", err)
			http.Error(w, "Erro interno ao buscar o filme", http.StatusInternalServerError)
		}
		return
	}

	// 4. Escrever a resposta de sucesso
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// @Summary      Deleta um filme por ID
// @Description  Remove um filme da coleção com base no seu ID.
// @Tags         Filmes
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "ID do Filme a ser deletado"
// @Success      204  "Filme deletado com sucesso (sem conteúdo de resposta)"
// @Failure      404  {object}  object{error=string} "Filme não encontrado"
// @Failure      500  {object}  object{error=string} "Erro interno no servidor"
// @Router       /movies/{id} [delete]
func (h *handler) deleteMovie(w http.ResponseWriter, r *http.Request) {
	log.Println("Requisição recebida: DELETE /movies/{id}")

	// 1. Extrair o ID da URL
	vars := mux.Vars(r)
	id := vars["id"]

	// 2. Chamar o serviço gRPC
	req := &pb.DeleteMovieRequest{Id: id}
	_, err := h.client.DeleteMovie(r.Context(), req) // A resposta de sucesso é vazia, por isso usamos '_'
	if err != nil {
		// 3. Traduzir o erro do gRPC para um erro HTTP
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			// Se o filme a ser deletado não for encontrado, retornamos 404
			http.Error(w, st.Message(), http.StatusNotFound)
		} else {
			log.Printf("Erro ao chamar DeleteMovie via gRPC: %v", err)
			http.Error(w, "Erro interno ao deletar o filme", http.StatusInternalServerError)
		}
		return
	}

	// 4. Escrever a resposta de sucesso
	// Para uma exclusão bem-sucedida, a convenção é retornar o status 204 No Content,
	// que significa "Eu fiz o que você pediu, mas não tenho nada para te mostrar em resposta".
	w.WriteHeader(http.StatusNoContent)
}
