pipeline {
    agent { label 'slave' }
    options {
        ansiColor('xterm')
        timeout(time: 20, unit: 'MINUTES')
        timestamps()
        buildDiscarder(logRotator(
            daysToKeepStr: '60',
            numToKeepStr: '300',
            artifactDaysToKeepStr: '5',
            artifactNumToKeepStr: '20'))
    }
    triggers {
        pollSCM('')
        issueCommentTrigger('.*test this please.*')
    }
    stages {
        stage('Setup') {
            steps {
                echo 'Setup'
                setBuildNameToGitCommit()
            }
        }
        stage('Run tests') {
            steps {
                runTests()
            }
        }
        stage('Build and push Docker image') {
            when {
                branch 'master'
            }
            steps {
                echo 'Build and push Docker image'
                vintedDockerImageBuildAndPublish('app')
            }
        }
        stage('Deploy to Kubernetes') {
            when {
                branch 'master'
            }
            steps {
                build job: 'services/kubernetes/cd-repo-apply-v2',
                    wait: true,
                    propagate: true,
                    parameters: [
                        string(name: 'REPOSITORY_SERVICE', value: 'app'),
                        string(name: 'OVERLAY', value: 'sandbox1'),
                        string(name: 'DOCKER_IMAGE_NAME', value: 'app'),
                        string(name: 'IMAGE_TAG', value: vintedDockerImageTag())
                    ]
            }
        }        
    }
    post {
        always {
            stopTestContainers()
        }
    }
}

void runTests() {
    vintedDockerImageBuildAndPublish('app', lintOnly: true)
    sh 'docker-compose build'
}

void stopTestContainers() {
    sh 'docker-compose down --rmi local --volumes'
}
