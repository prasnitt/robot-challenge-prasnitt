# Robot Warehouse System

A comprehensive REST API system for controlling a warehouse robot with command sequencing, task management, and real-time status monitoring.

![Build Status](https://github.com/prasnitt/robot-challenge-prasnitt/workflows/Robot%20Warehouse%20CI/badge.svg)

## 📋 Overview

This implementation provides a complete solution for the robot warehouse challenge with the following key features:

### **🎯 Solution Approach**
- **RESTful API Design**: Built using Go Gin framework for high-performance HTTP routing
- **Real-Time WebSocket Updates**: Provides live task status notifications via WebSocket connections
- **Concurrent Task Processing**: Implements goroutines and channels for non-blocking task execution
- **Thread-Safe Operations**: Uses mutexes to ensure data consistency across concurrent operations
- **Robust State Management**: Comprehensive task lifecycle with states (Pending → InProgress → Completed/Canceled/Aborted)
- **Event-Driven Architecture**: Uses channels to publish task state changes to connected clients
- **Boundary Validation**: Prevents robot from moving outside the 10x10 warehouse grid
- **Graceful Task Cancellation**: Supports real-time task cancellation even during execution

### **📊 System Architecture Diagram**

*The system follows a clean architecture pattern with separation of concerns between API handlers, business logic, and state management.*
```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              Robot Warehouse System                             │
│                                                                                 │
│  ┌──────────────────┐    ┌─────────────────┐    ┌─────────────────────────────┐ │
│  │   REST API       │    │   Robot Service │    │      Warehouse Grid         │ │
│  │   (Gin Router)   │    │                 │    │        (10x10)              │ │
│  │                  │    │                 │    │ (0,10)             (10,10)  │ │
│  │ POST /tasks      │◄──►│ EnqueueTask()   │    │  ┌─┬─┬─┬─┬─┬─┬─┬─┬─┬─┐      │ │
│  │ GET /state       │    │ CancelTask()    │    │  │ │ │ │ │ │ │ │ │ │ │      │ │
│  │ PUT /cancel      │    │ CurrentState()  │    │  ├─┼─┼─┼─┼─┼─┼─┼─┼─┼─┤      │ │
│  │                  │    │ ExecuteTask()   │    │  │ │ │ │ │ │ │ │ │ │ │      │ │
│  │ Swagger UI       │    │                 │    │  ├─┼─┼─┼─┼─┼─┼─┼─┼─┼─┤      │ │
│  └──────────────────┘    │ ┌─────────────┐ │    │  │ │ │ │ │🤖 │ │ │ │ │      │ │
│                          │ │ Task Queue  │ │    │  ├─┼─┼─┼─┼─┼─┼─┼─┼─┼─┤      │ │
│  ┌──────────────────┐    │ │(Channel)    │ │    │  │ │ │ │ │ │ │ │ │ │ │      │ │
│  │   WebSocket      │    │ │Max: 100     │ │    │  ├─┼─┼─┼─┼─┼─┼─┼─┼─┼─┤      │ │
│  │   Endpoint       │    │ └─────────────┘ │    │  │ │ │ │ │ │ │ │ │ │ │      │ │
│  │                  │◄───┤                 │    │  └─┴─┴─┴─┴─┴─┴─┴─┴─┴─┘      │ │
│  │ /api/v1/robot/   │    │ ┌─────────────┐ │    │  (0,0)              (10,0)  │ │
│  │ events           │    │ │Service State│ │    └─────────────────────────────┘ │
│  │                  │    │ │- Robot Pos  │ │                                    │
│  │ Real-time Events │◄───┤ │- Tasks Map  │ │    ┌─────────────────────────────┐ │
│  │ - Task Status    │    │ │- Task Count │ │    │                             │ │
│  │ - State Changes  │    │ │- Event Ch.  │ │    │                             │ │
│  │ - Timestamps     │    │ └─────────────┘ │    │                             │ │
│  └──────────────────┘    └─────────────────┘    │      Task State Flow        │ │
│  ┌──────────────────────────────────────────┐   │                             │ │
│  │              Concurrency                 │   │      Pending                │ │
│  │                                          │   │       │   │                 │ │
│  │ • Goroutines for task processing         │   │       │   └─► Canceled      │ │
│  │ • Mutex for thread-safe operations       │   │       ▼                     │ │
│  │ • Context for graceful shutdown          │   │  InProgress                 │ │
│  │ • Channels for task communication        │   │    │    │    │              │ │
│  │ • Event channels for WebSocket updates   │   │    │    │    └─► Aborted    │ │
│  │ • Real-time event broadcasting           │   │    │    │                   │ │
│  └──────────────────────────────────────────┘   │    │    │                   │ │
│                                                 │    │    │                   │ │
│                                                 │    │    └─► RequestCancel.. │ │
│                                                 │    │             │          │ │
│                                                 │    │             ▼          │ │
│                                                 │    ▼          Canceled      │ │
│                                                 │ Completed                   │ │
│                                                 └─────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────┘
```
---

### **🔄 System Architecture & Code Flow**

```
📥 HTTP Request → 🎯 Gin Router → 🤖 Robot Service → 📊 State Management → 🏭 Task Execution
                                          ↓
                                 🔄 Event Channel → 📡 WebSocket → 💻 Real-time Client Updates
```

1. **API Layer**: Handles HTTP requests with validation and response formatting
2. **Service Layer**: Manages robot logic, task queuing, and state transitions
3. **Task Processing**: Executes commands sequentially with configurable delays
4. **Event Broadcasting**: Publishes task state changes via event channels
5. **WebSocket Layer**: Delivers real-time updates to connected clients
6. **State Persistence**: Maintains robot position and task status in memory
7. **Concurrency Control**: Uses channels for task queuing and mutexes for data safety

### **🧪 Quality Assurance**
- **Comprehensive Unit Tests**: 88.9% code coverage with 50+ test cases
- **API Integration Tests**: Complete REST endpoint testing with mock services
- **Concurrent Testing**: Validates thread safety and race condition handling
- **Swagger Documentation**: Interactive API documentation with examples

## 📦 Packages Used

| Package | Version | Purpose | Description |
|---------|---------|---------|-------------|
| [Gin](https://github.com/gin-gonic/gin) | v1.10.1 | Web Framework | High-performance HTTP web framework |
| [Gorilla WebSocket](https://github.com/gorilla/websocket) | v1.5.3 | Real-time Communication | WebSocket implementation for live task updates |
| [Swaggo](https://github.com/swaggo/swag) | v1.16.5 | API Documentation | Swagger documentation generator |
| [Gin-Swagger](https://github.com/swaggo/gin-swagger) | v1.6.0 | Swagger UI | Swagger UI middleware for Gin |
| [UUID](https://github.com/google/uuid) | v1.6.0 | ID Generation | Unique task identifier generation |
| Go Standard Library | - | Core Logic | Context, sync, time, testing packages |

---

## 🚀 Setup & Run

### **Prerequisites**
- Go 1.24+ installed on your system
- Git for version control

### **Installation Steps**

1. **Clone the repository**
   ```bash
   git clone https://github.com/prasnitt/robot-challenge-prasnitt.git
   cd robot-challenge-prasnitt
   ```

2. **Install dependencies**
   ```bash
   go mod download
   go mod tidy
   ```

3. **Generate Swagger documentation** (Optional)
   ```bash
   swag init
   ```

4. **Run unit tests** (optional but recommended)
   ```bash
   go test ./... -v
   ```

5. **Start the application**
   ```bash
   go run main.go
   ```

6. **Access the system**
   - **API Base URL**: `http://localhost:8080/api/v1`
   - **Swagger UI**: `http://localhost:8080/swagger/index.html`
   - **Interactive API Testing**: Use Swagger UI to test all endpoints

### **📝 Usage Instructions**

#### **REST API Testing**
1. **Open Swagger UI** at `http://localhost:8080/swagger/index.html`
2. **Test API endpoints** using the interactive interface
3. **Monitor logs** in the terminal to see task execution flow
4. **Check robot movement** by calling the `/robot/state` endpoint

#### **WebSocket Real-Time Updates**
For real-time task status notifications, connect to the WebSocket endpoint:

**WebSocket URL**: `ws://localhost:8080/api/v1/robot/events`

**Using wscat (install with `npm install -g wscat`):**
```bash
# Connect to WebSocket for live task updates
wscat -c ws://localhost:8080/api/v1/robot/events

# You'll receive JSON messages like:
# {"task_id":"fdceaccc-5a27-4d9a-a17f-524c264f1741","state":"Pending","timestamp":"2025-07-20T00:24:47.638253+12:00"}

# {"task_id":"fdceaccc-5a27-4d9a-a17f-524c264f1741","state":"InProgress","timestamp":"2025-07-20T00:24:47.6396285+12:00"}
```

**Testing Flow with WebSocket:**
1. Open terminal and connect: `wscat -c ws://localhost:8080/api/v1/robot/events`
2. In another terminal/browser, create a task via REST API
3. Watch real-time status updates in the WebSocket connection
4. Status updates will show: Pending → InProgress → Completed/Canceled

**Note**: Current WebSocket implementation supports single client connections. Multi-client broadcast is planned as a future enhancement.

---

## 📸 Screenshots

| API Endpoint | Screenshot | Description |
|--------------|------------|-------------|
| **GET /robot/state** | ![Robot at origin](screenshots/init-state.jpg) | Shows robot at origin (0,0) with empty task queue |
| **POST /robot/tasks** | ![Task creation](screenshots/create-a-task.jpg) | Create task with commands "N E S W" and 5s delay |
| **GET /robot/state** (after task) | ![State after task completion](screenshots/state-after-task-completion.jpg) | Shows robot moved back to origin with completed task |
| **PUT /robot/tasks/{id}/cancel** | ![Task cancellation](screenshots/cancel-task-example.jpg) | Demonstrates real-time task cancellation |
| **WebSocket Events** | ![real-time task status](screenshots/realtime_update_on_websocket.jpg) | Real-time task status updates  about task creation -> InProgress -> Cancellation|

### **Suggested API Testing Flows**

1. **Basic Flow**: Create task → Check state → Watch logs for execution
2. **Real-Time Flow**: Connect WebSocket → Create task → Watch live status updates
3. **Cancellation Flow**: Create long task → Cancel mid-execution → Verify cancellation via WebSocket
4. **Boundary Testing**: Create task that exceeds warehouse limits → See validation error
5. **Multiple Tasks**: Create several tasks → Observe sequential processing via WebSocket events

---

## 🔮 Future Improvements

### **WebSocket & Real-Time Enhancements**
1. **Multi-Client WebSocket Hub**: Implement broadcast system to support multiple simultaneous WebSocket connections for task status updates


### **Configuration Management**
4. **Environment Variables**: Replace hardcoded values (port 8080, warehouse size 10) with configurable environment variables
5. **Config Files**: Support JSON/YAML configuration files for deployment flexibility

### **Scalability Enhancements**
6. **Infinite Task Queue**: Replace limited channel (100 tasks) with persistent queue (e.g. RabbitMQ/Database)
7. **Database Integration**: Store robot state and task history in PostgreSQL/MongoDB for persistence across restarts
8. **Horizontal Scaling**: Support multiple service instances with load balancing and shared state

### **Advanced Robot Intelligence**
9. **Path Optimization**: Instead of step-by-step commands, provide destination coordinates and let robot calculate optimal path
10. **Obstacle Avoidance**: Implement collision detection and dynamic path recalculation
11. **Multi-Robot Support**: Support multiple robots working simultaneously in the same warehouse

### **User Experience**
12. **Real-time Dashboard**: Web-based UI showing live robot position and task status with WebSocket integration
13. **Task Scheduling**: Support for delayed task execution and recurring tasks
14. **Event History**: Persistent logging of all task events and state changes with query capabilities

### **Alternative Event Mechanisms**
15. **Server-Sent Events (SSE)**: Alternative to WebSocket for one-way real-time updates with automatic reconnection
16. **Webhook Notifications**: HTTP callback support for external systems to receive task completion notifications
17. **Message Queue Integration**: Support for external message brokers (RabbitMQ, Apache Kafka) for enterprise-grade event distribution


---

## 🧩 API Endpoints

| Method | Endpoint | Description | Request Body | Response |
|--------|----------|-------------|--------------|----------|
| `GET` | `/api/v1/robot/state` | Get current robot state and tasks | None | `ServiceState` |
| `POST` | `/api/v1/robot/tasks` | Create new robot task | `AddTaskRequest` | `{task_id}` |
| `PUT` | `/api/v1/robot/tasks/{id}/cancel` | Cancel existing task | None | `{message}` |
| `WebSocket` | `/api/v1/robot/events` | Real-time task status updates | N/A | Task event stream |

### **WebSocket Event Format**
```json
{
  "task_id": "fdceaccc-5a27-4d9a-a17f-524c264f1741",
  "state": "InProgress",
  "timestamp": "2025-07-20T00:24:47.6396285+12:00"
}
```

**State Values**: `Pending`, `InProgress`, `Completed`, `Canceled`, `Aborted`, `RequestCancellation`

**Note**: WebSocket currently supports single client connections. Multi-client broadcast support is planned for future releases.

---

## 🏗️ Project Structure

```
robot-challenge-prasnitt/
├── docs/                     # Swagger documentation
├── internal/
│   ├── api/                  # HTTP handlers and routing
│   │   ├── handlers.go       # API endpoint handlers
│   │   ├── handlers_test.go  # API handler tests
│   │   └── routers.go        # Route configuration
│   └── robot/                # Core business logic
│       ├── command.go        # Robot command definitions
│       ├── service.go        # Main service implementation
│       ├── service_test.go   # Service unit tests
│       ├── state.go          # State management
│       ├── task.go           # Task creation and parsing
│       └── task_test.go      # Task unit tests
├── main.go                   # Application entry point
├── go.mod                    # Go module definition
└── README.md                 # Project documentation
```

---

# Problem Statement

We have installed a robot in our warehouse and now we need to send it commands to control it. We need you to implement the high level RESTful APIs, which can be called from a ground control station.

For convenience the robot moves along a grid in the roof of the warehouse and we have made sure that all of our warehouses are built so that the dimensions of the grid are 10 by 10. We've also made sure that all our warehouses are aligned along north-south and east-west axes. The robot also builds an internal x y coordinate map that aligns to the warehouse's physical dimensions. On the map, point (0, 0) indicates the most south-west and (10, 10) indicates the most north-east.

All of the commands to the robot consist of a single capital letter and different commands are delineated by whitespace.

The robot should accept the following commands:

- N move north
- W move west
- E move east
- S move south

**Example command sequences:**

The command sequence: "N E S W" will move the robot in a full square, returning it to where it started.

If the robot starts in the south-west corner of the warehouse then the following commands will move it to the middle of the warehouse.

"N E N E N E N E"

## Robot SDK Interface 

The robot provides a set of low level SDK functions in GO to control its movement. 

```go
type Warehouse interface {
	Robots() []Robot
}

type Robot interface {
	EnqueueTask(commands string) (taskID string, position chan RobotState, err chan error) 

	CancelTask(taskID string) error

	CurrentState() RobotState
}

type RobotState struct {
	X uint
	Y uint
	HasCrate bool
}
```

## Requirements
- Create a RESTful API to accept a series of commands to the robot. 
- Make sure that the robot doesn't try to move outside the warehouse.
- Create a RESTful API to report the command series's execution status.
- Create a RESTful API cancel the command series.
- The RESTful service should be written in Golang.

## Challenge
- The Robot SDK is still under development, you need to find a way to prove your API logic is working.
- The ground control station wants to be notified as soon as the command sequence completed. Please provide a high level design overview how you can achieve it. This overview is not expected to be hugely detailed but should clearly articulate the fundamental concept in your design.
