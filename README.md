PeerShare Backend
PeerShare is a decentralized file-sharing system built to leverage peer-to-peer (P2P) protocols for secure, efficient, and serverless file sharing. This backend project is implemented in Go, utilizing libp2p for P2P networking, and is designed to ensure scalability, resilience, and robust file transfers.

Peer Discovery: Enables peers to locate and connect with one another dynamically.
Distributed Hash Table (DHT): Implements SHA-based hashing for efficient file indexing and resource lookup.
Secure File Transfers: Ensures file integrity and confidentiality using cryptographic techniques.
Dockerized Environment: Supports simulations in isolated containers for reliable testing.
Scalability: Designed to handle an increasing number of peers and files seamlessly.
Technologies
This backend is built using the following technologies:

Go: Primary programming language.
libp2p: For peer-to-peer networking and communication.
Docker: To containerize the application for deployment and testing.
Git/GitHub: For version control and collaboration.
System Architecture
The PeerShare backend is structured as follows:

Peer Discovery Module:
Finds and connects with available peers in the network.
Distributed Hash Table (DHT):
Stores and retrieves file metadata for fast lookup.
File Transfer Protocol:
Handles file sending, receiving, and integrity validation.
Security:
Uses hashing algorithms (SHA-256, SHA-512) and public-key cryptography for secure connections.
Simulation Environment:
Dockerized setup to test the application in various network configurations.
Getting Started
Follow these steps to set up and run the backend:

Prerequisites
Go (v1.19 or higher): Install Go
Docker: Install Docker
Git: Install Git
