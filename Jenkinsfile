pipeline {
    agent any
    stages {
        stage('image') {
            steps {
                sh 'make image'
            }
        }
        stage('unit') {
            steps {
                sh 'make dockertest'
            }
        }
        stage('behave') {
            steps {
                sh 'make dockerbehave'
            }
        }
        stage('deploy') {
            steps {
                sh 'make deploy'
            }
        }
        stage('deploy test') {
            steps {
                sh 'make deploytest'
            }
        }
    }
}

