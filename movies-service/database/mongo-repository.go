// Local: movies-service/database/mongo_repository.go

package database

import (
	"context"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	// Importa nosso pacote de serviço para ter acesso à interface e ao modelo
	"github.com/alenrique/Movies-microservices/movies-service/service"
)

// mongoMovieRepository é a implementação do nosso repositório para o MongoDB.
type mongoMovieRepository struct {
	collection *mongo.Collection
}

// NewMongoMovieRepository é o construtor que cria uma nova instância do nosso repositório.
// Ele retorna a INTERFACE, e não a struct, para manter o acoplamento baixo.
func NewMongoMovieRepository(db *mongo.Database) service.MovieRepository {
	return &mongoMovieRepository{
		collection: db.Collection("movies"), // O nome da nossa collection no MongoDB
	}
}

// Save implementa o método de salvamento da interface MovieRepository.
func (r *mongoMovieRepository) Save(ctx context.Context, movie *service.Movie) error {
	_, err := r.collection.InsertOne(ctx, movie)
	return err
}

// FindByID implementa a busca por ID.
func (r *mongoMovieRepository) FindByID(ctx context.Context, id string) (*service.Movie, error) {
	var movie service.Movie

	// bson.M é um atalho para criar um filtro de busca
	err := r.collection.FindOne(ctx, bson.M{"id": id}).Decode(&movie)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Isso é importante para que o serviço saiba que o filme não foi encontrado
			return nil, nil
		}
		return nil, err
	}

	return &movie, nil
}

// FindAll implementa a busca por todos os documentos.
func (r *mongoMovieRepository) FindAll(ctx context.Context) ([]*service.Movie, error) {
	var movies []*service.Movie

	// Passamos um filtro vazio (bson.M{}) para pegar todos os documentos
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx) // Garante que o cursor seja fechado no final

	// Itera sobre os resultados do cursor
	for cursor.Next(ctx) {
		var movie service.Movie
		if err := cursor.Decode(&movie); err != nil {
			return nil, err
		}
		movies = append(movies, &movie)
	}

	return movies, nil
}

// DeleteByID implementa a exclusão por ID.
func (r *mongoMovieRepository) DeleteByID(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"id": id})
	return err
}

func (r *mongoMovieRepository) FindMaxID(ctx context.Context) (int, error) {
	// Busca todos os documentos da coleção.
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	maxID := 0
	// Itera sobre todos os filmes.
	for cursor.Next(ctx) {
		var movie struct {
			ID string `bson:"id"`
		}
		if err := cursor.Decode(&movie); err != nil {
			// Se um documento não puder ser decodificado, pulamos para o próximo.
			continue
		}

		// Converte o ID de string para int para podermos comparar.
		id, err := strconv.Atoi(movie.ID)
		if err != nil {
			// Se a conversão falhar (ex: o ID não é um número), pulamos.
			continue
		}

		// Verifica se o ID atual é o maior que já vimos.
		if id > maxID {
			maxID = id
		}
	}

	return maxID, nil
}
