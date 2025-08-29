// Local: movies-service/main.go

package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/bson" // Import corrigido
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"

	"github.com/alenrique/Movies-microservices/movies-service/database"
	"github.com/alenrique/Movies-microservices/movies-service/grpc_adapter"
	"github.com/alenrique/Movies-microservices/movies-service/service"
	pb "github.com/alenrique/Movies-microservices/proto"
)

func main() {
	// --- Conexão com o Banco de Dados ---
	log.Println("movies-service: Conectando ao MongoDB...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://mongodb:27017"))
	if err != nil {
		log.Fatalf("movies-service: Falha ao conectar com o MongoDB: %v", err)
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("movies-service: Falha ao pingar o MongoDB: %v", err)
	}
	log.Println("movies-service: Conectado ao MongoDB com sucesso!")

	// --- Injeção de Dependências ---
	movieRepo := database.NewMongoMovieRepository(client.Database("moviedb"))
	seedDatabase(ctx, client, movieRepo)
	movieService := service.NewMovieService(movieRepo)
	movieServer := grpc_adapter.NewGrpcMovieServer(movieService)

	// --- Configuração e Desligamento Gracioso do Servidor gRPC ---
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("movies-service: Falha ao escutar a rede: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterMovieServiceServer(grpcServer, movieServer)

	// Inicia o servidor em uma goroutine separada
	go func() {
		log.Printf("movies-service: Servidor gRPC escutando em %v", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("movies-service: Falha ao iniciar o servidor gRPC: %v", err)
		}
	}()

	// Canal para escutar por sinais de interrupção (Ctrl+C)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Bloqueia a execução até que um sinal seja recebido
	<-quit
	log.Println("movies-service: Sinal de desligamento recebido, parando o servidor gRPC...")

	// Tenta parar o servidor de forma graciosa
	grpcServer.GracefulStop()

	log.Println("movies-service: Servidor gRPC parado.")
}

// SUBSTITUA A FUNÇÃO ANTIGA POR ESTA
func seedDatabase(ctx context.Context, client *mongo.Client, movieRepo service.MovieRepository) {
	collection := client.Database("moviedb").Collection("movies")
	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Falha ao contar documentos: %v", err)
	}
	if count > 0 {
		log.Println("O banco de dados já contém dados. Pulo do 'seed'.")
		return
	}

	log.Println("Banco de dados vazio. Iniciando o 'seed' a partir de movies.json...")

	// 1. Definimos uma struct temporária que espelha o JSON exatamente
	type movieJSON struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
		Year  string `json:"year"`
	}

	file, err := os.Open("movies.json")
	if err != nil {
		log.Fatalf("Falha ao abrir movies.json: %v", err)
	}
	defer file.Close()

	// 2. Decodificamos o JSON para a nossa struct temporária
	var moviesFromJSON []movieJSON
	if err := json.NewDecoder(file).Decode(&moviesFromJSON); err != nil {
		log.Fatalf("Falha ao decodificar movies.json: %v", err)
	}

	// 3. Convertemos da struct temporária para a nossa struct de domínio final
	for _, m := range moviesFromJSON {
		year, _ := strconv.ParseInt(m.Year, 10, 32)

		movieToSave := &service.Movie{
			ID:       strconv.Itoa(m.ID), // Converte o ID de int para string
			Title:    m.Title,
			Director: "", // O JSON não tem diretor, então deixamos vazio
			Year:     int32(year),
		}

		err := movieRepo.Save(ctx, movieToSave)
		if err != nil {
			log.Printf("Falha ao inserir filme '%s': %v\n", movieToSave.Title, err)
		}
	}
	log.Println("Seed do banco de dados concluído com sucesso!")
}
