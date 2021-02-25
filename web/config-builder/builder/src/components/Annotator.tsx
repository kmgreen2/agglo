import * as React from "react"
import {Condition, ConditionBuilder, UndefinedCondition, SetConditionFn, ConditionJSONProps} from "./Condition";
import {Process} from "./Processes"
import {ChangeEvent} from "react";
import {Button, Form, Tab, Tabs} from "react-bootstrap";
import {MainModal} from "./MainModal";

interface AnnotatorBuilderProps {
    annotator?: Annotator
    setProcessState: (process: Process) => (void);
    validateProcessState: (process: Process, isEdit: boolean) => (void);

}
interface AnnotatorBuilderState {
    annotator: Annotator
    annotationTab: string
}

export class AnnotatorBuilder extends React.Component<AnnotatorBuilderProps, AnnotatorBuilderState> {
    isEdit: boolean

    constructor(props: AnnotatorBuilderProps){
        super(props);
        let annotator: Annotator | undefined = this.props.annotator;
        this.isEdit = true;
        if (this.props.annotator == null) {
            this.isEdit = false;
            annotator = new Annotator({
                name: "",
                annotations: Array<Annotation>(),
            })
        }
        if (annotator) {
            this.state = {
                annotator: annotator,
                annotationTab: annotator.name + "0"
            }
        }
    }

    componentDidMount() {
    }

    onChange = (event: ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
        let targetIdAry = event.target.id.split(":");
        let targetIdField = targetIdAry[0];

        this.setState((prev: AnnotatorBuilderState) => {
            let annotator = prev.annotator.copy();
            if (targetIdAry.length === 1) {
                if (targetIdField === "annotatorName") {
                    annotator.name = event.target.value
                }
            } else if (targetIdAry.length === 2) {
                let targetIdIdx = parseInt(targetIdAry[1]);
                let annotation = annotator.annotations[targetIdIdx].copy();
                if (targetIdField === "fieldKey") {
                    annotation.fieldKey = event.target.value;
                } else if (targetIdField === "fieldValue") {
                    annotation.fieldValue = event.target.value;
                }
                annotator.annotations[targetIdIdx] = annotation;
            }
            return {
                annotator: annotator
            }
        })
    }

    handleSave = (): any => {
        const errMessage = this.props.validateProcessState(this.state.annotator, this.isEdit);
        if (errMessage == null) {
            this.props.setProcessState(this.state.annotator);
        } else {
            return errMessage;
        }
        return null;
    }

    canAddCondition = (annotation: Annotation): boolean => {
        return annotation.fieldKey.length > 0 && annotation.fieldValue.length > 0;
    }

    saveConditionHandler = (annotation: Annotation, annotationIndex: number): SetConditionFn => {
        return (condition: Condition) => {
            let annotationCopy = annotation.copy();
            annotationCopy.condition = condition;
            this.setState((prev: AnnotatorBuilderState) => {
                let annotator = prev.annotator.copy();
                if (annotationIndex < annotator.annotations.length) {
                    annotator.annotations[annotationIndex] = annotationCopy;
                }
                return {
                    annotator: annotator
                }
            })
        }
    }

    deleteAnnotation = (annotationIndex: number) => {
        this.setState((prev: AnnotatorBuilderState) => {
            let annotator = prev.annotator.copy()
            let annotations = annotator.annotations.filter((elm: Annotation, idx: number) => {
                return idx !== annotationIndex;
            })
            annotator.annotations = annotations;
            return {
                annotator: annotator
            }
        }, () =>
            this.setAnnotationTab(this.state.annotator.name + `${this.state.annotator.annotations.length-1}`))
    }

    conditionForm = (annotation: Annotation, annotationIndex: number): (JSX.Element) => {
        return <>
            <ConditionBuilder saveConditionHandler={this.saveConditionHandler(annotation, annotationIndex)}
                              condition={annotation.condition}
                              enableButton={this.canAddCondition(annotation)} />
        </>
    }

    annotationsForm = (annotationIndex: number, annotation: Annotation): (JSX.Element) => {
        let defaultKey = annotation.fieldKey;
        let defaultValue = annotation.fieldValue;
        let conditionGroup = null;
        if (annotation.condition !== UndefinedCondition) {
            let condition = annotation.condition;
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
                <Form.Group controlId={`fieldKey:${annotationIndex}`}>
                    <Form.Label>Field Key</Form.Label>
                    <Form.Control type="text" onChange={this.onChange}
                                  defaultValue={defaultKey}/>
                    <Form.Text className="text-muted">
                        Field key for the annotation
                    </Form.Text>
                </Form.Group>
                <Form.Group controlId={`fieldValue:${annotationIndex}`}>
                    <Form.Label>Field Value</Form.Label>
                    <Form.Control type="text" onChange={this.onChange}
                                  defaultValue={defaultValue}/>
                    <Form.Text className="text-muted">
                        Field value for the annotation
                    </Form.Text>
                </Form.Group>
                {conditionGroup}
                <Form.Group>
                    <Button onClick={() => this.deleteAnnotation(annotationIndex)}>Delete</Button>
                    {this.conditionForm(annotation, annotationIndex)}
                </Form.Group>
            </Form>
        </>
    }

    appendAnnotation = () => {
        this.setState((prev: AnnotatorBuilderState) => {
            let annotator = prev.annotator.copy();
            let newAnnotation = new Annotation("", "", );
            annotator.annotations = annotator.annotations.concat(newAnnotation);
            return {
                annotator: annotator
            }
        }, () =>
            this.setAnnotationTab(this.state.annotator.name + `${this.state.annotator.annotations.length-1}`))
    }

    setAnnotationTab = (k: string|null) => {
        if (k) {
            this.setState({
                annotationTab: k
            })
        }
    }

    renderForm = (): (JSX.Element) => {
        const annotations = this.state.annotator.annotations;
        const annotatorKey = this.state.annotator.name;
        let annotationIndex = 0;
        const annotationTab = this.state.annotationTab;
        return <>
            <Form>
                <Form.Group controlId="annotatorName">
                    <Form.Label>Annotator Name</Form.Label>
                    <Form.Control type="text" onChange={this.onChange} defaultValue={annotatorKey}/>
                    <Form.Text className="text-muted"> Unique Annotator name </Form.Text>
                </Form.Group>
            </Form>
            <Tabs id="annotator-tabs" activeKey={annotationTab}
                  onSelect={this.setAnnotationTab}>
                {annotations.map(annotation => {
                    let title = "New Annotation";
                    if (annotation.fieldKey.length > 0 && annotation.fieldValue.length >0) {
                        title = `${annotation.fieldKey}:${annotation.fieldValue}`;
                    }
                    const tab = <Tab eventKey={annotatorKey + `${annotationIndex}`} title={title}>
                        {this.annotationsForm(annotationIndex, annotation)}
                    </Tab>
                    annotationIndex++;
                    return tab;
                })
                }
            </Tabs>
            <Button onClick={this.appendAnnotation}>Add Annotation</Button>
        </>
    }

    render() {
        let title = "Add Annotator";
        if (this.isEdit) {
            title = "Update Annotator";
        }
        return <>
            <MainModal title={title} renderForm={this.renderForm} handleSave={this.handleSave}/>
        </>
    }
}

interface AnnotationJSONProps {
    fieldKey: string
    value: string
    condition: ConditionJSONProps
}

class Annotation {
    fieldKey: string
    fieldValue: string
    condition: Condition

    constructor(fieldKey: string, fieldValue: string, condition?: Condition) {
        this.fieldKey = fieldKey;
        this.fieldValue = fieldValue;
        if (condition) {
            this.condition = condition;
        } else {
            this.condition = UndefinedCondition;
        }
    }

    toString = ():string => {
        return this.fieldKey + ":" + this.fieldValue + " if " + this.condition.toString();
    }

    static fromJSON = (props: AnnotationJSONProps): Annotation => {
        return new Annotation(props.fieldKey, props.value, Condition.fromJSON(props.condition));
    }

    toJSON = ():{} => {
        return {
            fieldKey: this.fieldKey,
            value: this.fieldValue,
            condition: this.condition.toJSON()
        }
    }

    copy = ():Annotation => {
        if (this.condition === UndefinedCondition) {
            return new Annotation(
                this.fieldKey,
                this.fieldValue,
                this.condition,
            )
        } else {
            return new Annotation(
                this.fieldKey,
                this.fieldValue,
                this.condition.copy(),
            )
        }
    }

    render() {
        return <>
            {this.toString()}
        </>
    }
}

export interface AnnotatorJSONProps {
    name: string
    annotations: Array<AnnotationJSONProps>
}

export interface AnnotatorProps {
    name: string
    annotations: Array<Annotation>
}
export interface AnnotatorState {}

export class Annotator extends React.Component<AnnotatorProps, AnnotatorState> {
    name: string
    annotations: Array<Annotation>

    constructor(props: AnnotatorProps) {
        super(props);
        this.name = props.name;
        this.annotations = props.annotations;
    }

    toString = ():string => {
        return JSON.stringify(this.toJSON());
    }

    toJSON = (): {} => {
        return {
            annotator: {
                name: this.name,
                annotations: this.annotations.map(a => a.toJSON())
            }
        }
    }

    static fromJSON = (props: AnnotatorJSONProps): Annotator => {
        return new Annotator({
            name: props.name,
            annotations: props.annotations.map(annotation => Annotation.fromJSON(annotation))
        });
    }

    copy = (): Annotator => {
        return new Annotator({
            name: this.name,
            annotations: Array.from(this.annotations),
        })
    }

    render() {
        return <>
            {this.name}
            </>
    }

}