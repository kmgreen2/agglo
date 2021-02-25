import * as React from "react";
import {Alert, Button, Modal} from 'react-bootstrap'

interface MainModalProps {
    renderForm: () => (JSX.Element)
    handleSave: () => (any)
    title: string
}

interface MainModalState {
    show: boolean
    showError: boolean
    errorMessage: string
}

export class MainModal extends React.Component<MainModalProps, MainModalState> {
    state: MainModalState

    constructor(props: MainModalProps) {
        super(props);
        this.state = {
            show: false,
            showError: false,
            errorMessage: "",
        }
    }

    showError = () => {
        this.setState({
            showError: true
        })
    }

    hideError = () => {
        this.setState({
            showError: false
        })
    }

    handleShow = () => {
        this.setState({
            show: true
        })
    }

    handleClose = () => {
        this.setState({
            show: false
        })
    }

    handleSave = () => {
        const result = this.props.handleSave();
        if (result) {
            this.setState({
                errorMessage: 'Error: ' + result
            })
            this.showError();
            return;
        }

        this.handleClose()
    }

    render() {
        return <>
            <Button variant="primary" onClick={this.handleShow}>
                {this.props.title}
            </Button>
            <Modal show={this.state.show} onHide={this.handleClose}>
                <Modal.Header closeButton>
                    <Modal.Title>{this.props.title}</Modal.Title>
                </Modal.Header>
                <Modal.Body>{this.props.renderForm()}</Modal.Body>
                <Modal.Footer>
                    <Alert show={this.state.showError} variant='danger' dismissible={true} onClose={this.hideError}>
                        {this.state.errorMessage}
                    </Alert>
                    <Button variant="secondary" onClick={this.handleClose}>
                        Close
                    </Button>
                    <Button variant="primary" onClick={this.handleSave}>
                        Save Changes
                    </Button>
                </Modal.Footer>
            </Modal>
            </>
    }
}