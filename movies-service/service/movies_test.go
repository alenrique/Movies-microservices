// Local: movies-service/service/movies_test.go

// Usamos o sufixo _test para indicar que este pacote é de teste.
// Isso nos força a testar apenas as funcionalidades públicas do pacote 'service'.
package service_test

import (
	"context"
	"strconv"
	"testing"

	// Importamos o pacote de serviço que queremos testar
	"github.com/alenrique/Movies-microservices/movies-service/service"
)

// --- 1. O Repositório Falso (Fake/Mock) ---
// fakeMovieRepository é uma implementação "dublê" da nossa interface MovieRepository.
// Ele usa um mapa em memória em vez de um banco de dados real.
type fakeMovieRepository struct {
	movies map[string]*service.Movie
}

// NewFakeMovieRepository cria uma nova instância do nosso repositório falso.
func NewFakeMovieRepository() *fakeMovieRepository {
	return &fakeMovieRepository{
		movies: make(map[string]*service.Movie),
	}
}

// Implementação dos métodos da interface para o nosso dublê:
func (f *fakeMovieRepository) Save(ctx context.Context, movie *service.Movie) error {
	f.movies[movie.ID] = movie
	return nil
}

func (f *fakeMovieRepository) FindMaxID(ctx context.Context) (int, error) {
	maxID := 0
	for _, movie := range f.movies {
		id, _ := strconv.Atoi(movie.ID)
		if id > maxID {
			maxID = id
		}
	}
	return maxID, nil
}

// (Implementações vazias para os outros métodos, pois não os usamos nestes testes)
func (f *fakeMovieRepository) FindByID(ctx context.Context, id string) (*service.Movie, error) {
	return nil, nil
}
func (f *fakeMovieRepository) FindAll(ctx context.Context) ([]*service.Movie, error) { return nil, nil }
func (f *fakeMovieRepository) DeleteByID(ctx context.Context, id string) error       { return nil }

// --- 2. Os Testes ---

// TestCreateMovie_Success testa o caminho feliz da criação de um filme.
func TestCreateMovie_Success(t *testing.T) {
	// Arrange (Preparação)
	repo := NewFakeMovieRepository()
	movieService := service.NewMovieService(repo)
	ctx := context.Background()
	movieToCreate := &service.Movie{Title: "The Matrix", Year: 1999}

	// Act (Ação)
	createdMovie, err := movieService.CreateMovie(ctx, movieToCreate)

	// Assert (Verificação)
	if err != nil {
		t.Errorf("Erro inesperado ao criar filme: %v", err)
	}
	if createdMovie.ID != "1" {
		t.Errorf("Esperava ID '1', mas recebeu '%s'", createdMovie.ID)
	}
	if len(repo.movies) != 1 {
		t.Errorf("Esperava que 1 filme fosse salvo no repositório, mas encontrou %d", len(repo.movies))
	}
}

// TestCreateMovie_FailsOnEmptyTitle testa se a validação de título vazio está funcionando.
func TestCreateMovie_FailsOnEmptyTitle(t *testing.T) {
	// Arrange
	repo := NewFakeMovieRepository()
	movieService := service.NewMovieService(repo)
	ctx := context.Background()
	movieToCreate := &service.Movie{Title: ""} // Título propositalmente vazio

	// Act
	_, err := movieService.CreateMovie(ctx, movieToCreate)

	// Assert
	if err == nil {
		t.Error("Esperava um erro ao criar filme com título vazio, mas não recebeu nenhum")
	}
}
