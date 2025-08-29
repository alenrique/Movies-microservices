// Local: movies-service/service/movies.go

package service

import (

	// Novo import
	"context"
	"errors"
	"strconv"
)

// === 1. Modelo de Domínio ===
// Esta é a estrutura de dados principal que nossa lógica de negócio vai usar.
// Ela representa um filme dentro do nosso sistema.
type Movie struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Director string `json:"director"`
	Year     int32  `json:"year"`
}

// === 2. Porta de Saída (Driven Port) ===
// Esta é a interface que define o que nossa aplicação PRECISA do mundo exterior.
// No caso, ela precisa de um meio para persistir e buscar dados de filmes.
// Qualquer banco de dados que queira se conectar ao nosso núcleo, DEVE implementar esta interface.
type MovieRepository interface {
	Save(ctx context.Context, movie *Movie) error
	FindByID(ctx context.Context, id string) (*Movie, error)
	FindAll(ctx context.Context) ([]*Movie, error)
	DeleteByID(ctx context.Context, id string) error
	FindMaxID(ctx context.Context) (int, error)
}

// === 3. Porta de Entrada (Driving Port) ===
// Esta interface define o que nossa aplicação OFERECE como funcionalidade.
// É a API pública do nosso núcleo de negócio.
type MovieService interface {
	CreateMovie(ctx context.Context, movie *Movie) (*Movie, error)
	GetMovie(ctx context.Context, id string) (*Movie, error)
	ListMovies(ctx context.Context) ([]*Movie, error)
	DeleteMovie(ctx context.Context, id string) error
}

// === 4. Implementação do Serviço (O Núcleo em si) ===
// Esta é a implementação concreta da nossa interface MovieService.
// Note que ela não sabe nada sobre MongoDB, apenas sobre a interface MovieRepository.
type movieService struct {
	repo MovieRepository // A única dependência é a nossa porta de saída.
}

// NewMovieService é um "construtor" que cria uma nova instância do nosso serviço.
// Ele recebe o adaptador de banco de dados (que implementa a interface Repository)
// e o injeta na nossa struct de serviço. Isso é Injeção de Dependência.
func NewMovieService(repo MovieRepository) MovieService {
	return &movieService{
		repo: repo,
	}
}

// Abaixo estão as implementações dos métodos da nossa lógica de negócio.
// Por enquanto, são apenas esqueletos chamando o repositório.
// Aqui é onde você adicionaria regras de negócio (validações, etc).

func (s *movieService) CreateMovie(ctx context.Context, movie *Movie) (*Movie, error) {
	if movie.Title == "" {
		return nil, errors.New("o título do filme não pode ser vazio")
	}

	// 1. Pergunta ao repositório qual é o maior ID existente.
	maxID, err := s.repo.FindMaxID(ctx)
	if err != nil {
		return nil, err
	}

	// 2. Calcula o novo ID e o converte para string.
	newID := maxID + 1
	movie.ID = strconv.Itoa(newID)

	// 3. Salva o filme com o novo ID numérico.
	err = s.repo.Save(ctx, movie)
	if err != nil {
		return nil, err
	}
	return movie, nil
}

func (s *movieService) GetMovie(ctx context.Context, id string) (*Movie, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *movieService) ListMovies(ctx context.Context) ([]*Movie, error) {
	return s.repo.FindAll(ctx)
}

func (s *movieService) DeleteMovie(ctx context.Context, id string) error {
	return s.repo.DeleteByID(ctx, id)
}
