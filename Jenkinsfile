pipeline {
    agent none
    stages {
        stage('Checkout') {
            agent any
            steps {
                checkout scm
                sh 'ls'
                sh 'echo $PATH'
            }
        }
        stage('Test') {
            agent any
            steps {
                sh 'cd test'
            }
            post {
                always {
                    junit 'test/*.xml'
                }
            }
        }
    }
}
