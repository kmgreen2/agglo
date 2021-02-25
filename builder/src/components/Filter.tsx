import * as React from "react"
import {Process} from "./Processes";
import {ChangeEvent} from "react";
import {Form} from "react-bootstrap";
import {MainModal} from "./MainModal";


interface FilterBuilderProps {
    filter?: Filter
    setProcessState: (process: Process) => (void);
    validateProcessState: (process: Process, isEdit: boolean) => (void);
}

interface FilterBuilderState {
    filter: Filter
}

export class FilterBuilder extends React.Component<FilterBuilderProps, FilterBuilderState> {
    isEdit: boolean

    constructor(props: FilterBuilderProps) {
        super(props);
        let filter: Filter | undefined = this.props.filter;
        this.isEdit = true;
        if (this.props.filter == null) {
            this.isEdit = false;
            filter = new Filter({
                name: "",
                filterSpec: new FilterSpec("", false)
            })
        }

        if (filter) {
            this.state = {
                filter: filter
            }
        }
    }

    componentDidMount() {
    }

    onChange = (event: ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
        this.setState((prev: FilterBuilderState) => {
            let filter = prev.filter.copy();
            if (event.target.id === "filterName") {
                filter.name = event.target.value;
            } else if (event.target.id === "regex") {
                filter.filterSpec.regex = event.target.value;
            } else if (event.target.id === "keepMatched") {
                filter.filterSpec.keepMatched = event.target.value === "on";
            }

            return {
                filter: filter
            }
        })
    }

    handleSave = (): any => {
        const errMessage = this.props.validateProcessState(this.state.filter, this.isEdit);
        if (errMessage == null) {
            this.props.setProcessState(this.state.filter);
        } else {
            return errMessage;
        }
        return null;
    }

    renderForm = (): (JSX.Element) => {
        const filterSpec = this.state.filter.filterSpec;
        const filterKey = this.state.filter.name;

        return <>
            <Form>
                <Form.Group controlId="filterName">
                    <Form.Label>Filter Name</Form.Label>
                    <Form.Control type="text" onChange={this.onChange} defaultValue={filterKey}/>
                    <Form.Text className="text-muted"> Unique Filter name </Form.Text>
                </Form.Group>
                <Form.Group controlId="regex">
                    <Form.Label>Regex</Form.Label>
                    <Form.Control type="text" onChange={this.onChange} defaultValue={filterSpec.regex}/>
                    <Form.Text className="text-muted"> Unique Filter name </Form.Text>
                </Form.Group>
                <Form.Group controlId="keepMatched">
                    <Form.Label>Keep Matched</Form.Label>
                    <Form.Check
                        custom
                        type="checkbox"
                        defaultChecked={filterSpec.keepMatched}
                        onChange={this.onChange}
                    />
                    <Form.Text className="text-muted">Keep regex matches (unchecked will filter them out)</Form.Text>
                </Form.Group>
            </Form>
        </>
    }

    render() {
        let title = "Add Filter";
        if (this.isEdit) {
            title = "Update Filter";
        }
        return <>
            <MainModal title={title} renderForm={this.renderForm} handleSave={this.handleSave}/>
        </>
    }
}

export class FilterSpec {
    regex: string
    keepMatched: boolean

    constructor(regex: string, keepMatched: boolean) {
        this.regex = regex;
        this.keepMatched = keepMatched;
    }

    copy = (): FilterSpec => {
        return new FilterSpec(this.regex, this.keepMatched);
    }
}

export interface FilterJSONProps {
    name: string
    regex: string
    keepMatched: boolean
}

interface FilterProps {
    name: string
    filterSpec: FilterSpec
}

interface FilterState {
}

export class Filter extends React.Component<FilterProps, FilterState> {
    name: string
    filterSpec: FilterSpec

    constructor(props: FilterProps) {
        super(props);
        this.name = props.name;
        this.filterSpec = props.filterSpec;
    }

    toString = ():string => {
        return JSON.stringify(this.toJSON());
    }

    static fromJSON = (props: FilterJSONProps): Filter => {
        return new Filter({
            name: props.name,
            filterSpec: new FilterSpec(
                props.regex,
                props.keepMatched
            )
        })
    }

    toJSON = (): {} => {
        return {
            filter: {
                name: this.name,
                regex: this.filterSpec.regex,
                keepMatched: this.filterSpec.keepMatched
            }
        }
    }

    copy = (): Filter => {
        return new Filter({
            name: this.name,
            filterSpec: this.filterSpec.copy()
        })
    }

    render() {
        return <>
            {this.name}
        </>
    }
}
