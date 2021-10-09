import * as React from "react"
import {
    Condition,
    ConditionBuilder,
    ConditionJSONProps,
    OperatorType,
    SetConditionFn,
    UndefinedCondition
} from "./Condition";
import { Map } from 'immutable'
import {Process} from "./Processes";
import {ChangeEvent} from "react";
import {Button, Form, Tab, Tabs} from "react-bootstrap";
import {MainModal} from "./MainModal";

export enum TransformationType {
    Unknown = "Unknown",
    Sum = "Sum",
    Copy = "Copy",
    MapRegex = "MapRegex",
    MapAdd = "MapAdd",
    MapMult = "MapMult",
    Count = "Count",
    LeftFold = "LeftFold",
    RightFold = "RightFold",
    Map = "Map",
    PopHead = "PopHead",
    PopTail = "PopTail"
}

export function GetTransformationTypes(): Array<string> {
    return Object.keys(TransformationType).filter((v) => v !== "Unknown")
}

export function StrToTransformationType(strTransformationType: string): TransformationType {
    return TransformationType[strTransformationType as keyof typeof TransformationType];
}

const TransformationArgNames = {
    Unknown: [],
    Sum: [],
    Copy: [],
    PopHead: [],
    PopTail: [],
    MapRegex: ["Regex", "Replace"],
    MapAdd: ["Const"],
    MapMult: ["Const"],
    Count: [],
    LeftFold: ["Path"],
    RightFold: ["Path"],
    Map: ["Path"],
}

interface TransformerBuilderProps {
    transformer?: Transformer
    setProcessState: (process: Process) => (void);
    validateProcessState: (process: Process, isEdit: boolean) => (void);
}

interface TransformerBuilderState {
    transformer: Transformer
    transformationTab: string
}

export class TransformerBuilder extends React.Component<TransformerBuilderProps, TransformerBuilderState> {
    isEdit: boolean

    constructor(props: TransformerBuilderProps) {
        super(props);
        let transformer: Transformer | undefined = this.props.transformer;
        this.isEdit = true;

        if (this.props.transformer == null) {
            this.isEdit = false;
            transformer = new Transformer({
                name: "",
                transformations: Array<Transformation>(),
                forwardInputFields: false
            })
        }

        if (transformer) {
            this.state = {
                transformer: transformer,
                transformationTab: transformer.name + "0"
            }
        }
    }

    componentDidMount() {
    }

    onChange = (event: ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
        let targetIdAry = event.target.id.split(":");
        let targetIdField = targetIdAry[0];

        this.setState((prev: TransformerBuilderState) => {
           let transformer = prev.transformer.copy();
           if (targetIdAry.length === 1) {
               if (targetIdField === "transformerName") {
                   transformer.name = event.target.value;
               } else if (targetIdField === "forwardInputFields") {
                   transformer.forwardInputFields = !transformer.forwardInputFields;
               }
           } else if (targetIdAry.length === 2) {
               let targetIdIdx = parseInt(targetIdAry[1]);
               let transformation = transformer.transformations[targetIdIdx].copy();
               if (targetIdField === "sourceField") {
                    transformation.sourceField = event.target.value;
               } else if (targetIdField === "targetField") {
                   transformation.targetField = event.target.value;
               } else if (targetIdField === "transformationType") {
                   transformation.transformationType = StrToTransformationType(event.target.value);
               }
               transformer.transformations[targetIdIdx] = transformation;
           } else if (targetIdAry.length === 3) {
               let targetArgsName = targetIdAry[1];
               let targetIdIdx = parseInt(targetIdAry[2]);
               let transformation = transformer.transformations[targetIdIdx].copy();
               transformation.transformationArgs =
                   transformation.transformationArgs.set(targetArgsName, event.target.value);
               transformer.transformations[targetIdIdx] = transformation;
           }
            return {
                transformer: transformer
            }
           // We want to call render() when the transformationType is changed, because each
           // type might render different arg fields
        }, () => targetIdField === "transformationType" && this.forceUpdate())
    }

    handleSave = (): any => {
        const errMessage = this.props.validateProcessState(this.state.transformer, this.isEdit);
        if (errMessage == null) {
            this.props.setProcessState(this.state.transformer);
        } else {
            return errMessage;
        }
        return null;
    }

    canAddCondition = (transformation: Transformation): boolean => {
        return (transformation.targetField.length === 0 && transformation.sourceField.length === 0 && transformation.transformationType === TransformationType.Copy) ||
            (transformation.targetField.length > 0 && transformation.sourceField.length > 0 &&
                transformation.transformationType !== TransformationType.Unknown)
    }

    saveConditionHandler = (transformation: Transformation, transformationIndex: number): SetConditionFn => {
        return (condition: Condition) => {
            let transformationCopy = transformation.copy();
            transformationCopy.condition = condition;
            this.setState((prev: TransformerBuilderState) => {
                let transformer = prev.transformer.copy();
                if (transformationIndex < transformer.transformations.length) {
                    transformer.transformations[transformationIndex] = transformationCopy;
                }
                return {
                    transformer: transformer
                }
            })
        }
    }

    setTransformationTab = (k: string|null) => {
        if (k) {
            this.setState({
                transformationTab: k
            })
        }
    }

    deleteTransformation = (transformationIndex: number) => {
        this.setState((prev: TransformerBuilderState) => {
            let transformer = prev.transformer.copy()
            transformer.transformations = transformer.transformations.filter((elm: Transformation, idx: number) => {
                return idx !== transformationIndex;
            })
            return {
                transformer: transformer
            }
        }, () =>
            this.setTransformationTab(this.state.transformer.name + `${this.state.transformer.transformations.length-1}`))
    }

    conditionForm = (transformation: Transformation, transformationIndex: number): (JSX.Element) => {
        return <>
            <ConditionBuilder saveConditionHandler={this.saveConditionHandler(transformation, transformationIndex)}
                              condition={transformation.condition}
                              enableButton={this.canAddCondition(transformation)} />
        </>
    }

    argsFormGroup = (transformation: Transformation, transformationIndex: number): (JSX.Element) => {
        if (!transformation.transformationType) {
            return <></>;
        }
        const form = Array.from(TransformationArgNames[transformation.transformationType]).map(argName => {
            return <Form.Group controlId={`arg:${argName}:${transformationIndex}`}>
                    <Form.Label>{argName}</Form.Label>
                    <Form.Control type={"text"} onChange={this.onChange}
                                  defaultValue={transformation.transformationArgs.get(argName)} />
                    <Form.Text className="text-muted">
                        {argName} argument for transformation
                    </Form.Text>
                </Form.Group>
        })
        return <>{form}</>;
    }

    transformationsForm = (transformation: Transformation, transformationIndex: number): (JSX.Element) => {
        let conditionGroup = null;
        let defaultChoice = "Choose...";
        let defaultTransformationType = defaultChoice;

        if (transformation.transformationType && transformation.transformationType !== TransformationType.Unknown) {
            defaultTransformationType = transformation.transformationType;
        }

        if (transformation.condition !== UndefinedCondition) {
            let condition = transformation.condition;
            conditionGroup = <Form.Group key={condition.toString()}>
                <Form.Label>Condition</Form.Label>
                <Form.Row>
                    <Condition operatorType={condition.operatorType}
                               operator={condition.operator}
                               lhs={condition.lhs}
                               rhs={condition.rhs} />
                </Form.Row>
            </Form.Group>
        }

        return <>
            <Form>
                <Form.Group controlId={`sourceField:${transformationIndex}`}>
                    <Form.Label>Source Field</Form.Label>
                    <Form.Control type={"text"} onChange={this.onChange} defaultValue={transformation.sourceField} />
                    <Form.Text className="text-muted">
                        Source field for transformation
                    </Form.Text>
                </Form.Group>
                <Form.Group controlId={`targetField:${transformationIndex}`}>
                    <Form.Label>Target Field</Form.Label>
                    <Form.Control type={"text"} onChange={this.onChange} defaultValue={transformation.targetField} />
                    <Form.Text className="text-muted">
                        Target field for transformation
                    </Form.Text>
                </Form.Group>
                <Form.Group controlId={`transformationType:${transformationIndex}`}>
                    <Form.Label>Transformation Type</Form.Label>
                    <Form.Control as="select" onChange={this.onChange} defaultValue={defaultTransformationType}>
                        <option key={defaultChoice}>{defaultChoice}</option>
                        {GetTransformationTypes().map(t => <option key={t}>{t}</option>)}
                    </Form.Control>
                    <Form.Text className="text-muted">
                        Choose a transformation type
                    </Form.Text>
                </Form.Group>
                {this.argsFormGroup(transformation, transformationIndex)}
                {conditionGroup}
                <Form.Group>
                    <Button onClick={() => this.deleteTransformation(transformationIndex)}>Delete</Button>
                    {this.conditionForm(transformation, transformationIndex)}
                </Form.Group>
            </Form>
        </>
    }

    appendTransformation = () => {
        this.setState((prev: TransformerBuilderState) => {
            let transformer = prev.transformer.copy();
            let newTransformation = new Transformation("", "", TransformationType.Unknown);
            transformer.transformations = transformer.transformations.concat(newTransformation);
            return {
                transformer: transformer
            }
        }, () =>
            this.setTransformationTab(this.state.transformer.name + `${this.state.transformer.transformations.length-1}`))
    }

    renderForm = (): (JSX.Element) => {
        const transformations = this.state.transformer.transformations;
        const transformerKey = this.state.transformer.name;
        let transformationIndex = 0;
        const transformationTab = this.state.transformationTab;

        return <>
            <Form>
                <Form.Group controlId="transformerName">
                    <Form.Label>Transformer Name</Form.Label>
                    <Form.Control type="text" onChange={this.onChange} defaultValue={transformerKey}/>
                    <Form.Text className="text-muted"> Unique Transformer name </Form.Text>
                </Form.Group>
                <Form.Group controlId={"forwardInputFields"}>
                    <Form.Label>Forward Input Fields</Form.Label>
                    <Form.Check
                        custom
                        type="checkbox"
                        defaultChecked={this.state.transformer.forwardInputFields}
                        onChange={this.onChange}
                    />
                    <Form.Text className="text-muted">Forward Input fields</Form.Text>
                </Form.Group>
            </Form>
            <Tabs id="transformer-tabs" activeKey={transformationTab}
                  onSelect={this.setTransformationTab}>
                {transformations.map(transformation => {
                    let title = "New Annotation";
                    if (transformation.sourceField && transformation.targetField) {
                        title = `${transformation.sourceField}:${transformation.targetField}`;
                    } else {
                        title = "all:all"
                    }
                    const tab = <Tab eventKey={transformerKey + `${transformationIndex}`} title={title}>
                        {this.transformationsForm(transformation, transformationIndex)}
                    </Tab>
                    transformationIndex++;
                    return tab;
                })
                }
            </Tabs>
            <Button onClick={this.appendTransformation}>Add Transformation</Button>
        </>
    }

    render() {
        let title = "Add Transformer";
        if (this.isEdit) {
            title = "Update Transformer";
        }
        return <>
            <MainModal title={title} renderForm={this.renderForm} handleSave={this.handleSave}/>
        </>
    }
}

function transformationArgsToSpec(transformationType: TransformationType, args: Map<string, string>): {} {
   if (transformationType === TransformationType.Map) {
       return {
            path: args.get('Path')
       }
   } else if (transformationType === TransformationType.MapAdd) {
       return {
            value: args.get('Const')
       }
   } else if (transformationType === TransformationType.MapMult) {
       return {
            value: args.get('Const')
       }
   } else if (transformationType === TransformationType.MapRegex) {
       return {
            regex: args.get('Regex'),
            replace: args.get('Replace')
       }
   } else if (transformationType === TransformationType.LeftFold) {
       return {
            path: args.get('Path')
       }
   } else if (transformationType === TransformationType.RightFold) {
       return {
           path: args.get('Path')
       }
   }
   return {}
}

function argsSpecToTransformationSpec(argsSpec: {}): Map<string, string> {
    let m = Map<string, string>();
    if ("path" in argsSpec) {
        m = m.set("Path", argsSpec["path"] as string);
    }
    if ("value" in argsSpec) {
        m = m.set("Const", argsSpec["value"] as string);
    }
    if ("regex" in argsSpec) {
        m = m.set("Regex", argsSpec["regex"] as string);
    }
    if ("replace" in argsSpec) {
        m = m.set("Replace", argsSpec["replace"] as string);
    }
    return m
}

interface TransformationJSONProps {
    sourceField: string
    targetField: string
    transformation: {
        transformationType: string
        condition: ConditionJSONProps
        mapRegexArgs: {}
        mapArgs: {}
        mapAddArgs: {}
        mapMultArgs: {}
        leftFoldArgs: {}
        rightFoldArgs: {}
    }
}

export class Transformation {
    sourceField: string
    targetField: string
    condition: Condition
    transformationType: TransformationType
    transformationArgs: Map<string, string>

    constructor(sourceField: string, targetField: string,
                transformationType: TransformationType, condition?: Condition,
                transformationArgs?: Map<string, string>) {
        this.sourceField = "";
        if (sourceField !== undefined) {
            this.sourceField = sourceField;
        }
        this.targetField = "";
        if (targetField !== undefined) {
            this.targetField = targetField;
        }
        this.transformationType = transformationType;
        if (condition) {
            this.condition = condition;
        } else {
            this.condition = UndefinedCondition;
        }
        if (transformationArgs) {
            this.transformationArgs = transformationArgs;
        } else {
            this.transformationArgs = Map<string, string>();
        }
    }

    static fromJSON = (props: TransformationJSONProps): Transformation => {
        let transformationArgs = Map<string, string>();
        let sourceField = "";
        let targetField = "";
        if (props.transformation.mapAddArgs) {
            transformationArgs = argsSpecToTransformationSpec(props.transformation.mapAddArgs);
        } else if (props.transformation.mapArgs) {
            transformationArgs = argsSpecToTransformationSpec(props.transformation.mapArgs);
        } else if (props.transformation.mapRegexArgs) {
            transformationArgs = argsSpecToTransformationSpec(props.transformation.mapRegexArgs);
        } else if (props.transformation.mapMultArgs) {
            transformationArgs = argsSpecToTransformationSpec(props.transformation.mapMultArgs);
        } else if (props.transformation.leftFoldArgs) {
            transformationArgs = argsSpecToTransformationSpec(props.transformation.leftFoldArgs);
        } else if (props.transformation.rightFoldArgs) {
            transformationArgs = argsSpecToTransformationSpec(props.transformation.rightFoldArgs);
        }

        if (props.transformation.transformationType === TransformationType.Copy && !props.sourceField && !props.targetField) {
            sourceField = "";
            targetField = "";
        } else {
            sourceField = props.sourceField;
            targetField = props.targetField;
        }


        return new Transformation(sourceField, targetField,
            StrToTransformationType(props.transformation.transformationType.replace("Transform", "")),
            Condition.fromJSON(props.transformation.condition),
            transformationArgs);
    }

    toJSON = (): {} => {
        let transformArgs = transformationArgsToSpec(this.transformationType, this.transformationArgs);

        if (this.transformationType === TransformationType.Map) {
            return {
                sourceField: this.sourceField,
                targetField: this.targetField,
                transformation: {
                    transformationType: 'Transform' + this.transformationType,
                    condition: this.condition.toJSON(),
                    mapArgs: transformArgs
                }
            }
        } else if (this.transformationType === TransformationType.MapAdd){
            return {
                sourceField: this.sourceField,
                targetField: this.targetField,
                transformation: {
                    transformationType: 'Transform' + this.transformationType,
                    condition: this.condition.toJSON(),
                    mapAddArgs: transformArgs
                }
            }
        } else if (this.transformationType === TransformationType.MapMult){
            return {
                sourceField: this.sourceField,
                targetField: this.targetField,
                transformation: {
                    transformationType: 'Transform' + this.transformationType,
                    condition: this.condition.toJSON(),
                    mapMultArgs: transformArgs
                }
            }
        } else if (this.transformationType === TransformationType.MapRegex){
            return {
                sourceField: this.sourceField,
                targetField: this.targetField,
                transformation: {
                    transformationType: 'Transform' + this.transformationType,
                    condition: this.condition.toJSON(),
                    mapRegexArgs: transformArgs
                }
            }
        } else if (this.transformationType === TransformationType.LeftFold){
            return {
                sourceField: this.sourceField,
                targetField: this.targetField,
                transformation: {
                    transformationType: 'Transform' + this.transformationType,
                    condition: this.condition.toJSON(),
                    leftFoldArgs: transformArgs
                }
            }
        } else if (this.transformationType === TransformationType.RightFold){
            return {
                sourceField: this.sourceField,
                targetField: this.targetField,
                transformation: {
                    transformationType: 'Transform' + this.transformationType,
                    condition: this.condition.toJSON(),
                    rightFoldArgs: transformArgs
                }
            }
        }

        return {
            sourceField: this.sourceField,
            targetField: this.targetField,
            transformation: {
                transformationType: 'Transform' + this.transformationType,
                condition: this.condition.toJSON(),
            }
        }
    }

    copy = (): Transformation => {
        if (this.condition !== UndefinedCondition) {
            return new Transformation(this.sourceField, this.targetField, this.transformationType,
                this.condition.copy(), this.transformationArgs)
        } else {
            return new Transformation(this.sourceField, this.targetField, this.transformationType,
                this.condition, this.transformationArgs)
        }
    }
}

export interface TransformerJSONProps {
    name: string
    specs: Array<TransformationJSONProps>
    forwardInputFields: boolean
}

export interface TransformerProps {
    name: string
    transformations: Array<Transformation>
    forwardInputFields: boolean
}

export interface TransformerState {
}

export class Transformer extends React.Component<TransformerProps, TransformerState> {
    name: string
    transformations: Array<Transformation>
    forwardInputFields: boolean

    constructor(props: TransformerProps) {
        super(props);
        this.name = props.name;
        this.transformations = props.transformations;
        this.forwardInputFields = props.forwardInputFields;
    }

    toString = ():string => {
        return JSON.stringify(this.toJSON());
    }

    static fromJSON = (props: TransformerJSONProps): Transformer => {
        return new Transformer({
            name: props.name,
            transformations: props.specs.map(tr => Transformation.fromJSON(tr)),
            forwardInputFields: props.forwardInputFields
        });
    }

    toJSON = (): {} => {
       return {
           transformer: {
               name: this.name,
               specs: this.transformations.map((t => t.toJSON())),
               forwardInputFields: this.forwardInputFields
           }
       }
    }

    copy = ():Transformer => {
        return new Transformer({
            name: this.name,
            transformations: Array.from(this.transformations),
            forwardInputFields: this.forwardInputFields
        })
    }

    render() {
        return <>
            {this.name}
        </>
    }
}
