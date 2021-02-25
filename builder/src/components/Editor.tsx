import * as React from "react"
import {External, ExternalProps} from "./Externals";
import {Process, ProcessJSONProps} from "./Processes";
import {Form} from "react-bootstrap";
import {Pipeline, PipelineJSONProps} from "./Pipelines";
import { saveAs } from 'file-saver';
import {ChangeEvent, RefObject} from "react";
import {Aggregator} from "./Aggregator";
import {Annotator} from "./Annotator";
import {Completer} from "./Completer";
import {Spawner} from "./Spawner";
import {Filter} from "./Filter";
import {Transformer} from "./Transformer";
import {Tee} from "./Tee";
import {Continuation} from "./Continuation";
import {Map} from "immutable";
import {Entwine} from "./Entwine";

export interface EditorProps {
    getMetadataState: () => Map<string, string>;
    setMetadataState: (key: string, value: string) => (void);
    getExternalState: () => Array<External>;
    getProcessState: () => Array<Process>;
    getPipelineState: () => Array<Pipeline>;
    setProcessState:(process: Process) => (void);
    validateProcessState: (process: Process, isEdit: boolean) => (void);
    deleteProcess: (process: Process)=>(void);
    setExternalState: (external: External)=>(void);
    deleteExternal: (external: External)=>(void);
    validateExternalState: (external: External, isEdit: boolean) => any;
    validatePipelineState: (pipeline: Pipeline, isEdit: boolean) => any;
    setPipelineState: (pipeline: Pipeline) => (void);
    deletePipeline: (pipeline: Pipeline) => (void);
}
export interface EditorState {
}

export class Editor extends React.Component<EditorProps, EditorState> {
    textInputRef: RefObject<HTMLTextAreaElement>
    constructor(props: EditorProps) {
        super(props);
        this.textInputRef = React.createRef<HTMLTextAreaElement>();
    }

    updateState = (jsonText: string) => {
        let pipelines = Array<Pipeline>();
        let processes = Array<Process>();
        let externals = Array<External>();

        const configJSON = JSON.parse(jsonText);

        if ("partitionUuid" in configJSON) {
            const partitionUuid = configJSON["partitionUuid"] as string;
            this.props.setMetadataState("partitionUuid", partitionUuid);
        }
        if ("externalSystems" in configJSON) {
            const externalSystemsJson = configJSON["externalSystems"] as Array<ExternalProps>;
            externals = externalSystemsJson.map(ext => External.fromJSON(ext));
        }
        if ("processDefinitions" in configJSON) {
            const processDefinitionsJson = configJSON["processDefinitions"] as Array<ProcessJSONProps>;
            processes = processDefinitionsJson.map(proc => {
                if ("aggregator" in proc) {
                    return Aggregator.fromJSON(proc["aggregator"]);
                } else if ("annotator" in proc) {
                    return Annotator.fromJSON(proc["annotator"]);
                } else if ("completer" in proc) {
                    return Completer.fromJSON(proc["completer"]);
                } else if ("filter" in proc) {
                    return Filter.fromJSON(proc["filter"]);
                } else if ("spawner" in proc) {
                    return Spawner.fromJSON(proc["spawner"]);
                } else if ("tee" in proc) {
                    return Tee.fromJSON(proc["tee"]);
                } else if ("transformer" in proc) {
                    return Transformer.fromJSON(proc["transformer"]);
                } else if ("continuation" in proc) {
                    return Continuation.fromJSON(proc["continuation"]);
                } else if ("entwine" in proc) {
                    return Entwine.fromJSON(proc["entwine"]);
                }
                throw Error("cannot find process definition type");
            })
        }
        if ("pipelines" in configJSON) {
            const pipelinesJson = configJSON["pipelines"] as Array<PipelineJSONProps>;
            pipelines = pipelinesJson.map(pipeline => Pipeline.fromJSON(pipeline));
        }

        this.props.getPipelineState().forEach((pipeline, idx, ary) => {
            this.props.deletePipeline(pipeline);
        })

        this.props.getProcessState().forEach((process, idx, ary) => {
            this.props.deleteProcess(process);
        })

        this.props.getExternalState().forEach((external, idx, ary) => {
            this.props.deleteExternal(external);
        })

        externals.forEach((external, idx, ary) => {
            this.props.setExternalState(external);
        })

        processes.forEach((process, idx, ary) => {
            this.props.setProcessState(process);
        })

        pipelines.forEach((pipeline, idx, ary) => {
            this.props.setPipelineState(pipeline);
        })

    }

    handleChange = () => {
        if (this.textInputRef.current) {
            this.updateState(this.textInputRef.current.value);
        }
    }

    handleReset = () => {
        this.updateState("{}");
    }

    download = (data: string) => {
        const blob = new Blob([data], {type: "application/json;charset=utf-8"});
        saveAs(blob);
    }

    upload = () => {
        // @ts-ignore
        const file = document.getElementById('uploadFile01').files[0];
        const reader = new FileReader();
        reader.addEventListener('load', (event) => {
            if (event.target) {
                if (event.target.result && typeof(event.target.result) === "string") {
                    this.updateState(event.target.result);
                }
            }
        })
        reader.readAsBinaryString(file);
    }

    render() {
        let metadata = this.props.getMetadataState().get("partitionUuid");
        if (metadata == null) {
            metadata = "";
        }
        const jsonText = {
            partitionUuid: metadata,
            pipelines: this.props.getPipelineState().map(pipeline => pipeline.toJSON()),
            externalSystems: this.props.getExternalState().map(ext => ext.toJSON()),
            processDefinitions: this.props.getProcessState().map(proc => proc.toJSON())
        }
        return <>
            <Form>
                <Form.Group controlId="editorText">
                    <Form.Control as="textarea" style={{ height: "400px" }}
                                  ref={this.textInputRef}>
                        {JSON.stringify(jsonText, null, 2)}
                    </Form.Control>
                </Form.Group>
            </Form>
            <div className="input-group mb-3 w-50">
                <div className="custom-file">
                    <input type="file" className="custom-file-input" id="uploadFile01"
                           aria-describedby="inputGroupFile01" onChange={this.upload} autoFocus={true}/>
                        <label className="custom-file-label" htmlFor="inputGroupFile01">Choose file to upload</label>
                </div>
                <div className="input-group-append">
                    <button className="btn btn-outline-secondary" type="button" id="inputGroupFile02"
                            onClick={() => this.download(JSON.stringify(jsonText, null, 2))}>
                        Download Config
                    </button>
                    <button className="btn btn-outline-secondary" type="button" id="inputGroupFile03"
                            onClick={() => this.handleChange()}>
                        Save Config
                    </button>
                    <button className="btn btn-outline-secondary" type="button" id="inputGroupFile04"
                            onClick={() => this.handleReset()}>
                        Reset Config
                    </button>
                </div>
            </div>
        </>
    }
}
