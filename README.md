# Texas Hold'em Hand Evaluator & Probability Calculator (GKE)

This project implements a Texas Hold'em hand evaluation and Monte Carlo equity calculator, deployed on Google Kubernetes Engine (GKE).

## Technology Stack
- Backend: Go (REST API)
- Frontend: Flutter Web (Dart)
- Containerization: Docker
- Cloud: Google Kubernetes Engine (GKE)
- CI/CD: Google Cloud Build

## Features
- Evaluate best poker hand from 2 hole cards + 5 community cards
- Determine winner between two hands
- Monte Carlo simulation to estimate winning probability with configurable number of players and simulations
- Web UI for card selection and result visualization

## Repository Structure
backend/ # Go REST API (hand evaluation + Monte Carlo simulation)
frontend/ # Flutter Web UI
Dockerfile # Dockerfiles for backend & frontend


## API Endpoints (Backend)
- `POST /api/evaluate` – Evaluate best hand
- `POST /api/equity` – Monte Carlo equity simulation

> The backend is intended to be called by the frontend UI.

## Live Deployment (GKE)
Frontend (public URL):  

http://34.136.74.178/


Backend (example endpoint):  


http://104.197.178.51:8080/api/evaluate


> Note: The external IPs may change if the cluster is redeployed.

## How to Run Locally

### Backend
```bash
cd backend
go run ./cmd/server

Frontend
cd frontend
flutter run -d chrome

Deployment to GKE (Summary)
gcloud config set project temppoker
gcloud services enable container.googleapis.com artifactregistry.googleapis.com cloudbuild.googleapis.com

gcloud artifacts repositories create temppoker --repository-format=docker --location=us-central1

gcloud container clusters create temppoker-cluster --zone us-central1-a --num-nodes 2

gcloud builds submit --tag us-central1-docker.pkg.dev/temppoker/temppoker/backend
kubectl create deployment temppoker-backend --image=us-central1-docker.pkg.dev/temppoker/temppoker/backend
kubectl expose deployment temppoker-backend --type=LoadBalancer --port 8080

gcloud builds submit --tag us-central1-docker.pkg.dev/temppoker/temppoker/frontend
kubectl create deployment temppoker-frontend --image=us-central1-docker.pkg.dev/temppoker/temppoker/frontend
kubectl expose deployment temppoker-frontend --type=LoadBalancer --port 80