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
                parallel(
                    "Runtime":{
                        withEnv(["GOPATH=${WORKSPACE}", "TESTDIR=${WORKSPACE}/${PROJ_PATH}/test"]){
                            sh 'cd ${TESTDIR}; ginkgo --focus="Run*" -v'
                        }
                    },
                    "K8s":{
                        withEnv(["GOPATH=${WORKSPACE}", "TESTDIR=${WORKSPACE}/${PROJ_PATH}/test"]){
                            sh 'cd ${TESTDIR}; ginkgo --focus="K8s*" -v'
                        }
                    },
                )
            }
            post {
                always {
                    junit 'test/*.xml'
                    sh 'cd test/; vagrant destroy -f'
                }
            }
        }
    }
}
