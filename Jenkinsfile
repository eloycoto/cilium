pipeline {
    agent none
    environment {
        PROJ_PATH = "src/github.com/cilium/cilium"
    }

    options {
        timeout(time: 120, unit: 'MINUTES')
        timestamps()
    }

    stages {
        stage('Checkout') {
            agent any
            steps {
                sh 'rm -rf src; mkdir -p src/github.com/cilium'
                sh 'ln -s $WORKSPACE src/github.com/cilium/cilium'
                checkout scm
                sh 'echo "" > test/ssh-config'
            }
        }
        stage('Test') {
            agent any
            environment {
                GOPATH="${WORKSPACE}"
                TESTDIR="${WORKSPACE}/${PROJ_PATH}/test"
            }
            steps {
                parallel(
                    "Runtime":{
                        sh 'cd ${TESTDIR}; ginkgo --focus="Run*" -v -noColor'
                    },
                    "K8s":{
                        sh 'cd ${TESTDIR}; ginkgo --focus="K8s*" -v -noColor'
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
