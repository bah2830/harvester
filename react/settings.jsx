class Settings extends React.Component {
    submit(e) {
        e.preventDefault();
    }

    save() {
        var settings = {
            jira: {
                url: document.getElementById('jiraUrl').value,
                user: document.getElementById('jiraUser').value,
                pass: document.getElementById('jiraPass').value
            },
            harvest: {
                user: document.getElementById('harvestUser').value,
                pass: document.getElementById('harvestPass').value
            }
        }
        external.invoke(JSON.stringify(settings));
    }


    render() {
        return (
            <form onSubmit={this.submit}>
                <h4>Jira</h4>
                <div className="form-group row">
                        <label htmlFor="jiraUrl" className="col-4 col-form-label">URL</label>
                        <div className="col">
                        <input
                            type="text"
                            className="form-control"
                            id="jiraUrl"
                            placeholder="url"
                            defaultValue={settings.data.jira.url}
                        />
                        </div>
                    </div>
                <div className="form-group row">
                    <label htmlFor="jiraUser" className="col-4 col-form-label">Username</label>
                    <div className="col">
                        <input
                            type="text"
                            className="form-control"
                            id="jiraUser"
                            placeholder="username"
                            defaultValue={settings.data.jira.user}
                        />
                    </div>
                </div>
                <div className="form-group row">
                    <label htmlFor="jiraPass" className="col-4 col-form-label">Password</label>
                    <div className="col">
                        <input
                            type="password"
                            className="form-control"
                            id="jiraPass"
                            placeholder="password"
                            defaultValue={settings.data.jira.pass}
                        />
                    </div>
                </div>

                <br /><br />

                <h4>Harvest</h4>
                <div className="form-group row">
                    <label htmlFor="harvestUser" className="col-4 col-form-label">Username</label>
                    <div className="col">
                        <input
                            type="text"
                            className="form-control"
                            id="harvestUser"
                            placeholder="username"
                            defaultValue={settings.data.harvest.user}
                        />
                    </div>
                </div>
                <div className="form-group row">
                    <label htmlFor="harvestPass" className="col-4 col-form-label">Password</label>
                    <div className="col">
                        <input
                            type="password"
                            className="form-control"
                            id="harvestPass"
                            placeholder="password"
                            defaultValue={settings.data.harvest.pass}
                        />
                    </div>
                </div>

                <br />
                <button id="save" className="btn btn-primary btn-block" onClick={this.save}>Save</button>
            </form>
        );
    }
}

if (typeof settings !== 'undefined') {
    const render = () =>  ReactDOM.render(<Settings />, document.getElementById('app'));
    settings.render = render;
    render();
}