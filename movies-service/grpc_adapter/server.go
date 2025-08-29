// Local: movies-service/grpc_adapter/server.go

package grpc_adapter

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	// Importa os pacotes gerados e o nosso serviço
	"github.com/alenrique/Movies-microservices/movies-service/service"
	pb "github.com/alenrique/Movies-microservices/proto"
)

// GrpcMovieServer é a nossa implementação do servidor gRPC.
// Ele satisfaz a interface gerada pelo protoc.
type GrpcMovieServer struct {
	pb.UnimplementedMovieServiceServer // Incorporação obrigatória para compatibilidade
	service                            service.MovieService
}

// NewGrpcMovieServer é o construtor para nosso servidor gRPC.
func NewGrpcMovieServer(svc service.MovieService) *GrpcMovieServer {
	return &GrpcMovieServer{service: svc}
}

// CreateMovie é a implementação do método gRPC para criar um filme.
func (s *GrpcMovieServer) CreateMovie(ctx context.Context, req *pb.CreateMovieRequest) (*pb.Movie, error) {
	// --- ESTE É O PADRÃO ADAPTER ---

	// 1. Traduzir: Converte a requisição gRPC para o nosso modelo de domínio interno.
	domainMovie := &service.Movie{
		Title:    req.GetTitle(),
		Director: req.GetDirector(),
		Year:     req.GetYear(),
	}

	// 2. Chamar o Núcleo: Executa a lógica de negócio real.
	// O adaptador não sabe COMO o filme é criado, ele apenas delega para o serviço.
	createdMovie, err := s.service.CreateMovie(ctx, domainMovie)
	if err != nil {
		// Em uma aplicação real, você traduziria o erro para um status code gRPC apropriado.
		return nil, err
	}

	// 3. Traduzir de Volta: Converte o resultado do nosso domínio para a resposta gRPC.
	return &pb.Movie{
		Id:       createdMovie.ID, // Nota: Ainda precisamos gerar o ID no nosso service!
		Title:    createdMovie.Title,
		Director: createdMovie.Director,
		Year:     createdMovie.Year,
	}, nil
}

// TODO: Implementar os outros métodos: GetMovie, ListMovies e DeleteMovie.

// ListMovies implementa o método gRPC para listar todos os filmes.
func (s *GrpcMovieServer) ListMovies(ctx context.Context, req *pb.ListMoviesRequest) (*pb.ListMoviesResponse, error) {
	// --- PADRÃO ADAPTER PARA LISTAS ---

	// 1. Chamar o Núcleo: A requisição de entrada (req) é vazia, então não há nada para traduzir.
	// Chamamos diretamente nossa lógica de negócio para buscar a lista de filmes.
	domainMovies, err := s.service.ListMovies(ctx)
	if err != nil {
		return nil, err // Traduzir para um status code gRPC
	}

	// 2. Traduzir a Saída: Esta é a parte principal. O serviço nos deu uma lista
	// no formato do nosso domínio ([]*service.Movie). Precisamos convertê-la para o
	// formato gRPC ([]*pb.Movie).
	var grpcMovies []*pb.Movie

	// Usamos um loop 'for' para percorrer cada filme do domínio.
	for _, domainMovie := range domainMovies {
		// Para cada um, criamos um novo filme no formato gRPC e copiamos os dados.
		grpcMovie := &pb.Movie{
			Id:       domainMovie.ID,
			Title:    domainMovie.Title,
			Director: domainMovie.Director,
			Year:     domainMovie.Year,
		}
		// Adicionamos o filme convertido à nossa lista gRPC.
		grpcMovies = append(grpcMovies, grpcMovie)
	}

	// 3. Retornar a Resposta gRPC: O nosso .proto define que a resposta é um
	// objeto ListMoviesResponse, que por sua ve contém uma lista de filmes.
	return &pb.ListMoviesResponse{
		Movies: grpcMovies,
	}, nil
}

// GetMovie implementa o método gRPC para buscar um filme por ID.
func (s *GrpcMovieServer) GetMovie(ctx context.Context, req *pb.GetMovieRequest) (*pb.Movie, error) {
	// 1. Extrair o Parâmetro: Pegamos o ID da requisição gRPC.
	movieID := req.GetId()
	if movieID == "" {
		// É uma boa prática validar a entrada. Se o ID estiver vazio, retornamos um erro claro.
		return nil, status.Errorf(codes.InvalidArgument, "O ID do filme não pode ser vazio")
	}

	// 2. Chamar o Núcleo: Passamos o ID para a nossa lógica de negócio.
	domainMovie, err := s.service.GetMovie(ctx, movieID)
	if err != nil {
		// Se houver um erro do banco de dados (ex: conexão caiu), repassamos o erro.
		return nil, status.Errorf(codes.Internal, "Erro interno ao buscar o filme: %v", err)
	}

	// 3. Lidar com o "Não Encontrado": Este é o caso especial.
	// O nosso repositório retorna (nil, nil) se o filme não for encontrado.
	// O serviço repassa isso. Aqui, nós traduzimos essa resposta para um erro gRPC padrão.
	if domainMovie == nil {
		return nil, status.Errorf(codes.NotFound, "Filme com o ID '%s' não encontrado", movieID)
	}

	// 4. Traduzir a Saída: Se encontramos o filme, convertemos do nosso formato de domínio
	// para o formato de resposta gRPC, como já fizemos antes.
	return &pb.Movie{
		Id:       domainMovie.ID,
		Title:    domainMovie.Title,
		Director: domainMovie.Director,
		Year:     domainMovie.Year,
	}, nil
}

// DeleteMovie implementa o método gRPC para deletar um filme por ID.
func (s *GrpcMovieServer) DeleteMovie(ctx context.Context, req *pb.DeleteMovieRequest) (*pb.DeleteMovieResponse, error) {
	// 1. Extrair e Validar o Parâmetro
	movieID := req.GetId()
	if movieID == "" {
		return nil, status.Errorf(codes.InvalidArgument, "O ID do filme não pode ser vazio")
	}

	// 2. Chamar o Núcleo
	err := s.service.DeleteMovie(ctx, movieID)
	if err != nil {
		// Aqui poderíamos ter uma lógica para checar se o erro foi porque o filme não existia
		// e retornar um 'NotFound', ou se foi um erro interno. Por simplicidade,
		// trataremos qualquer erro do serviço como um erro interno por enquanto.
		return nil, status.Errorf(codes.Internal, "Erro interno ao deletar o filme: %v", err)
	}

	// 3. Retornar a Resposta de Sucesso
	// Conforme definido no nosso .proto, a resposta para um delete bem-sucedido
	// é uma mensagem vazia. Apenas retornamos a struct de resposta vazia.
	return &pb.DeleteMovieResponse{}, nil
}
