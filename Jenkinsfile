pipeline {
    agent {
        node {
            label 'multimarket-test'
        }
    }

    environment {
        GOPATH = "${WORKSPACE}"
        GOBIN = "${WORKSPACE}/bin"
        GOPROXY='https://goproxy.cn,direct'
        APP_NAME = 'event-pod-services'
        CONTAINER_PORT = '8081'
        INTERNAL_PORT = '8080'
        REGISTRY = 'registry.cn-hangzhou.aliyuncs.com'  // Optional: image registry URL
    }

    options {
        timeout(time: 30, unit: 'MINUTES')
        disableConcurrentBuilds()
        buildDiscarder(logRotator(numToKeepStr: '10'))
    }

    stages {
        stage('Check ENV') {
            steps {
                echo '=== Checking build environment ==='
                sh 'go version'
            }
        }

        stage('Dependencies') {
            steps {
                echo '=== Installing dependencies ==='
                  sh 'go mod download'
            }
        }        

        stage('Build') {
            steps {
                echo '=== Building BIN ==='
                sh 'make'
            }
        }

        stage('Build Image') {
            steps {
                echo '=== Building Docker image ==='
                script {
                    def imageTag = "${env.BUILD_NUMBER}"
                    sh """
                        docker build -t ${APP_NAME}:${imageTag} .
                    """
                }
            }
        }

        stage('Deploy') {
            steps {
                echo '=== Deploying application ==='
                script {
                    def imageTag = "${env.BUILD_NUMBER}"
                    sh """
                        # Stop and remove old container if exists
                        docker stop ${APP_NAME} 2>/dev/null || true
                        docker rm ${APP_NAME} 2>/dev/null || true
                        
                        # Start new container
                        docker run -d \\
                            --name ${APP_NAME} \\
                            --restart unless-stopped \\
                            -p ${CONTAINER_PORT}:${INTERNAL_PORT} \\
                            ${APP_NAME}:${imageTag}
                        
                        # Wait for container startup
                        sleep 5
                        
                        echo "=== Deploy success! URL: http://localhost:${CONTAINER_PORT} ==="
                    """
                }
            }
        }
    }

    post {
        success {
            echo 'Build and deploy SUCCESS!'
        }
        failure {
            echo 'Build and deploy FAILED!'
        }
        always {
            cleanWs()
        }
    }
}
