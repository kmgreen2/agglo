import * as React from "react";
import {MainModal} from "./MainModal"
import {Button, Card, Form, ListGroup, ListGroupItem, Table} from "react-bootstrap"
import {ChangeEvent} from "react";
import * as protobufjs from "protobufjs";
import {Root} from "protobufjs";

export enum ExternalType {
    Unknown = "Unknown",
    KVStore = "KVStore",
    ObjectStore = "ObjectStore",
    PubSub = "PubSub",
    Http = "Http",
    LocalFile = "LocalFile",
}

export function StrToExternalType(strExternalType: string): ExternalType {
    return ExternalType[strExternalType as keyof typeof ExternalType];
}

export function ExternalTypeToStr(externalType: ExternalType): string {
    return externalType
}

export function GetExternalTypes(): Array<string> {
    return Object.keys(ExternalType).filter((v) => v !== "Unknown")
}

interface ExternalBuilderProps {
    external?: External
    setExternalState: (external: External) => (void);
    validateExternalState: (external: External, isEdit: boolean) => any;
}

interface ExternalBuilderState {
    external: External
}

// Form-based builder for an external provider
class ExternalBuilder extends React.Component<ExternalBuilderProps, ExternalBuilderState> {
    isEdit: boolean

    constructor(props: ExternalBuilderProps) {
        super(props);
        let external: External | undefined = this.props.external;
        this.isEdit = true;
        if (this.props.external == null) {
            this.isEdit = false;
            external = new External({
                name: "",
                externalType: ExternalType.KVStore,
                connectionString: "",
            });
        }
        if (external) {
            this.state = {
                external: external
            }
        }
    }

    componentDidMount() {
    }

    onChange = (event: ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
        let external = this.state.external.copy()
        if (event.target.id === "externalName") {
            external.name = event.target.value;
        } else if (event.target.id === "externalType") {
            external.externalType = StrToExternalType(event.target.value);
        } else if (event.target.id === "connectionString"){
            external.connectionString = event.target.value;
        } else {
            return;
        }
        this.setState({
            external: external
        })
    }

    handleSave = (): any => {
        const errMessage = this.props.validateExternalState(this.state.external, this.isEdit);
        if (errMessage == null) {
            this.props.setExternalState(this.state.external);
        } else {
            return errMessage;
        }
        return null;
    }

    renderForm = (): (JSX.Element) => {
        const externalName = this.state.external.name;
        const externalType = this.state.external.externalType;
        const connectionString = this.state.external.connectionString;

        return <>
            <Form>
                <Form.Group controlId="externalName">
                    <Form.Label>Name</Form.Label>
                    <Form.Control type="text" onChange={this.onChange} defaultValue={externalName}
                                  readOnly={this.isEdit}/>
                    <Form.Text className="text-muted">
                        Choose a unique name for this external provider
                    </Form.Text>
                </Form.Group>
                <Form.Group controlId="externalType">
                    <Form.Label>External Provider Type</Form.Label>
                    <Form.Control as="select" onChange={this.onChange} defaultValue={StrToExternalType(externalType)}>
                        {GetExternalTypes().map(t => <option key={t}>{t}</option>)}
                    </Form.Control>
                    <Form.Text className="text-muted">
                        Choose an external provider type
                    </Form.Text>
                </Form.Group>
                <Form.Group controlId="connectionString">
                    <Form.Label>Connection String</Form.Label>
                    <Form.Control type="text" onChange={this.onChange} defaultValue={connectionString}/>
                    <Form.Text className="text-muted">
                        Provider-specific connection string
                    </Form.Text>
                </Form.Group>
            </Form>
        </>

    }

    // Render a button that allows you to create a new external provider or edit an existing one
    render() {
        let title = "Add External";
        if (this.isEdit) {
            title = "Update External";
        }
        return <>
            <MainModal title={title} renderForm={this.renderForm} handleSave={this.handleSave}/>
        </>
    }
}

export interface ExternalProps {
    name: string;
    externalType: ExternalType;
    connectionString: string;
}

interface ExternalState {
}

// Container for an external provider
export class External extends React.Component<ExternalProps, ExternalState> {
    name: string;
    externalType: ExternalType;
    connectionString: string;

    constructor(props: ExternalProps) {
        super(props);
        this.name = props.name;
        this.externalType = props.externalType;
        this.connectionString = props.connectionString;
    }

    toString = ():string => {
        return this.name + ':' + this.externalType + ':' + this.connectionString;
    }

    copy = ():External => {
        return new External({
            name: this.name,
            externalType: this.externalType,
            connectionString: this.connectionString,
        })
    }

    toJSON = (): {} => {
        return {
            name: this.name,
            connectionString: this.connectionString,
            externalType: "External" + this.externalType,
        }
    }

    static fromJSON = (props: ExternalProps):External => {
        if (props.externalType.startsWith("External")) {
            return new External({
                name: props.name,
                externalType: StrToExternalType(props.externalType.replace("External", "")),
                connectionString: props.connectionString
            })
        }
        return new External(props);
    }

    // Render a summary of this external provider
    render() {
        return <>
            <div>
                <p><b>{this.name}</b> ({this.externalType})</p>
            </div>
        </>
    }
}

export interface ExternalsProps {
    setExternalState: (external: External)=>(void);
    deleteExternal: (external: External)=>(void);
    validateExternalState: (external: External, isEdit: boolean) => any;
    getExternalState: ()=> Array<External>;
}

interface ExternalsState {}

// Main page for externals
export class Externals extends React.Component<ExternalsProps, ExternalsState> {
    // Builders *need* a new instance each time, so we generate a
    //new key for each one
    builderId: number

    constructor(props: ExternalsProps) {
        super(props);
        this.builderId = 0;
    }

    // Render the current external providers, with edit/delete and the ability
    // to add new external providers
    render() {
        return <>
            <ListGroup>
                {this.props.getExternalState().map((external) =>
                    <ListGroupItem>
                        <ListGroup horizontal={true}>
                            <ListGroupItem>
                                <External key={external.name} name={external.name} connectionString={external.connectionString}
                                          externalType={external.externalType} />
                            </ListGroupItem>
                            <ListGroupItem>
                                <ExternalBuilder key={external.name + "-builder"} setExternalState={this.props.setExternalState}
                                                 validateExternalState={this.props.validateExternalState}
                                                 external={external}/>
                            </ListGroupItem>
                            <ListGroupItem>
                                <Button onClick={() => this.props.deleteExternal(external)}>Delete</Button>
                            </ListGroupItem>
                        </ListGroup>
                    </ListGroupItem>
                )}
            </ListGroup>
            <ExternalBuilder key={this.builderId++} setExternalState={this.props.setExternalState}
                             validateExternalState={this.props.validateExternalState} />
        </>
    }
}
