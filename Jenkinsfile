pipeline {
    agent none
    environment {
        PROJ_PATH = "src/github.com/cilium/cilium"
    }
    stages {
        stage('Checkout') {
            agent any
            steps {
                sh 'rm -rf src; mkdir -p src/github.com/cilium'
                sh 'ln -s $WORKSPACE src/github.com/cilium/cilium'
                checkout scm
            }
        }
        stage('Test') {
            agent any
            steps {
                withEnv(["GOPATH=${WORKSPACE}", "TESTDIR=${WORKSPACE}/${PROJ_PATH}/test"]){
                    sh 'ls $WORKSPACE/src/github.com/cilium/'
                    sh 'echo ${TESTDIR}'
                    sh 'echo ${PROJ_PATH}'
                    sh 'cd ${TESTDIR}; ginkgo --focus="K8s*"'
                }
            }
            post {
                always {
                    junit 'test/*.xml'
                }
            }
        }
    }
}
