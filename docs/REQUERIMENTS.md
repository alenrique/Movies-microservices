# Análise de Requisitos - Sistema de Gerenciamento de Filmes

**Data:** 28 de Agosto de 2025
**Autor:** Henrique Alencar

## 1. Visão Geral do Projeto

O objetivo deste projeto é desenvolver um sistema de microsserviços para realizar operações de CRUD (Criar, Ler, Atualizar, Deletar) em uma coleção de filmes. O sistema é composto por uma API REST pública (API Gateway) que se comunica com um serviço de domínio interno (Serviço de Filmes) via gRPC. A persistência dos dados é realizada em um banco de dados MongoDB.

O projeto é totalmente containerizado com Docker e orquestrado com Docker Compose para permitir a inicialização do ambiente completo com um único comando, demonstrando proficiência em arquiteturas de software modernas, limpas e escaláveis.

## 2. Histórias de Usuário (User Stories)

Como o "usuário" deste sistema é um cliente de API (outro serviço, um frontend, etc.), as histórias são contadas sob essa perspectiva para focar no valor entregue pela API.

| ID   | História de Usuário                                                                                                                                                            |
| :--- | :----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| HU01 | **Como um** cliente da API, **eu quero** enviar os dados de um novo filme para `POST /movies`, **para que** ele seja armazenado e eu receba os dados do filme criado com seu novo ID. |
| HU02 | **Como um** cliente da API, **eu quero** fazer uma requisição `GET /movies`, **para que** eu receba uma lista de todos os filmes cadastrados no sistema.                          |
| HU03 | **Como um** cliente da API, **eu quero** fazer uma requisição `GET /movies/{id}`, **para que** eu receba os dados completos do filme correspondente àquele ID.                      |
| HU04 | **Como um** cliente da API, **eu quero** fazer uma requisição `DELETE /movies/{id}`, **para que** o filme correspondente seja removido do sistema.                               |

## 3. Requisitos Funcionais (RF)

| ID   | Descrição                                                                                               |
| :--- | :------------------------------------------------------------------------------------------------------ |
| RF01 | O sistema deve expor um endpoint `POST /movies` para permitir a criação de um novo filme.                 |
| RF02 | O sistema deve expor um endpoint `GET /movies` para permitir a listagem de todos os filmes cadastrados.     |
| RF03 | O sistema deve expor um endpoint `GET /movies/{id}` para permitir a busca de um filme específico por seu ID. |
| RF04 | O sistema deve expor um endpoint `DELETE /movies/{id}` para permitir a exclusão de um filme por seu ID.     |
| RF05 | Na primeira inicialização com um banco de dados vazio, o sistema deve populá-lo com os dados de `movies.json`. |
| RF06 | O sistema deve gerar um ID numérico incremental e único para cada novo filme criado.                       |

## 4. Requisitos Não-Funcionais (RNF)

| ID    | Categoria             | Descrição                                                                                             |
| :---- | :-------------------- | :---------------------------------------------------------------------------------------------------- |
| RNF01 | Linguagem             | A aplicação deve ser desenvolvida integralmente em Go.                                                |
| RNF02 | Arquitetura de Código | A lógica de negócio deve ser isolada das dependências externas utilizando a Arquitetura Hexagonal.       |
| RNF03 | Arquitetura de Sistema| O sistema deve ser estruturado em microsserviços, com separação clara entre o API Gateway e o serviço de domínio. |
| RNF04 | Comunicação Interna   | A comunicação entre o API Gateway e o Serviço de Filmes deve ser feita utilizando o protocolo gRPC/Protobuf. |
| RNF05 | Persistência          | O banco de dados utilizado para armazenar os dados dos filmes deve ser o MongoDB.                     |
| RNF06 | Containerização       | Todos os componentes da aplicação (microsserviços e banco de dados) devem ser containerizados com Docker. |
| RNF07 | Orquestração          | O ambiente completo deve ser inicializável com um único comando através do Docker Compose.             |
| RNF08 | Documentação da API   | A API REST pública deve ser documentada utilizando a especificação Swagger/OpenAPI.                  |
| RNF09 | Qualidade de Código   | O núcleo de negócio do Serviço de Filmes deve possuir testes unitários para garantir seu funcionamento.     |
| RNF10 | Controle de Versão    | O código-fonte final deve ser entregue em um repositório Git (GitHub).                              |
| RNF11 | Usabilidade           | O projeto deve incluir um arquivo `README.md` com instruções claras para compilação e execução.         |