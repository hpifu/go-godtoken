pipeline {
    agent any
    stages {
        stage('unit') {
            steps {
                sh 'make dockertest'
            }
        }
        stage('image') {
            steps {
                sh 'make image'
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

