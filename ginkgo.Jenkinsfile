def failFast = { String branch ->
  if (branch == "origin/master" || branch == "master") {
    return '--failFast=false'
  } else {
    return '--failFast=true'
  }
}

def myparams = params.collect{
    string(name: it.key, value: it.value)
}

pipeline {
    agent {
        label 'baremetal'
    }
    environment {
        PROJ_PATH = "src/github.com/cilium/cilium"
        TESTDIR = "${WORKSPACE}/${PROJ_PATH}/"
        MEMORY = "3072"
    }

    options {
        timeout(time: 120, unit: 'MINUTES')
        timestamps()
    }
    stages {
        stage('Checkout') {
            steps {
                sh 'env'
                sh 'rm -rf src; mkdir -p src/github.com/cilium'
                sh 'ln -s $WORKSPACE src/github.com/cilium/cilium'
                checkout scm
                build job: 'Test_projects/eloy-test', parameters: myparams
            }
        }
        stage('UnitTesting') {
            environment {
                GOPATH="${WORKSPACE}"
                TESTDIR="${WORKSPACE}/${PROJ_PATH}/"
            }
            steps {
                sh "cd ${TESTDIR}; make tests-ginkgo"
            }
            post {
                always {
                    sh "cd ${TESTDIR}; make clean-ginkgo-tests || true"
                }
            }
        }
        stage('Boot VMs'){
            environment {
                TESTDIR="${WORKSPACE}/${PROJ_PATH}/test"
            }
            steps {
                sh 'cd ${TESTDIR}; K8S_VERSION=1.7 vagrant up --no-provision'
                sh 'cd ${TESTDIR}; K8S_VERSION=1.10 vagrant up --no-provision'
            }
        }
        stage('BDD-Test-PR') {
            environment {
                GOPATH="${WORKSPACE}"
                TESTDIR="${WORKSPACE}/${PROJ_PATH}/test"
                FAILFAST = failFast(env.GIT_BRANCH)
            }
            options {
                timeout(time: 90, unit: 'MINUTES')
            }
            steps {
                parallel(
                    "Runtime":{
                        sh 'cd ${TESTDIR}; ginkgo --focus=" RuntimeValidated*" -v -noColor ${FAILFAST}'
                    },
                    "K8s-1.7":{
                        sh 'cd ${TESTDIR}; K8S_VERSION=1.7 ginkgo --focus=" K8sValidated*" -v -noColor ${FAILFAST}'
                    },
                    "K8s-1.10":{
                        sh 'cd ${TESTDIR}; K8S_VERSION=1.10 ginkgo --focus=" K8sValidated*" -v -noColor ${FAILFAST}'
                    },
                    failFast: true
                )
            }
            post {
                always {
                    junit 'test/*.xml'
                    // Temporary workaround to test cleanup
                    // rm -rf ${GOPATH}/src/github.com/cilium/cilium
                    sh 'cd test/; ./post_build_agent.sh || true'
                    sh 'cd test/; ./archive_test_results.sh || true'
                    archiveArtifacts artifacts: "test_results_${JOB_BASE_NAME}_${BUILD_NUMBER}.zip", allowEmptyArchive: true
                }
            }
        }
    }
    post {
        always {
            sh 'cd ${TESTDIR}/test/; K8S_VERSION=1.7 vagrant destroy -f || true'
            sh 'cd ${TESTDIR}/test/; K8S_VERSION=1.10 vagrant destroy -f || true'
            cleanWs()
        }
    }
}
