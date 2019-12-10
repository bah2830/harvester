import React from 'react';

export class Settings extends React.Component {
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
        astilectron.sendMessage('settings|' + JSON.stringify(settings));
    }

    description(options) {
        return <small id={options.id + 'Help'} className="form-text text-muted">{options.description}</small>;
    }

    render() {
        const forms = [
            {
                'group': 'Jira',
                'forms': [
                    {
                        'label': 'URL',
                        'type': 'text',
                        'id': 'jiraUrl',
                        'placeholder': 'url',
                        'defaultValue': (appData.data.settings.jira && appData.data.settings.jira.url)
                    },
                    {
                        'label': 'Username',
                        'type': 'text',
                        'id': 'jiraUser',
                        'placeholder': 'username',
                        'defaultValue': (appData.data.settings.jira && appData.data.settings.jira.user)
                    },
                    {
                        'label': 'Password',
                        'type': 'password',
                        'id': 'jiraPass',
                        'placeholder': 'password',
                        'defaultValue': (appData.data.settings.jira && appData.data.settings.jira.pass)
                    },
                ]
            },
            {
                'group': 'Harvest',
                'forms': [
                    {
                        'label': 'Account Id',
                        'type': 'text',
                        'id': 'harvestUser',
                        'placeholder': 'account_id',
                        'defaultValue': (appData.data.settings.harvest && appData.data.settings.harvest.user),
                        'description': 'A new application can be created at https://id.getharvest.com/developers'
                    },
                    {
                        'label': 'Token',
                        'type': 'password',
                        'id': 'harvestPass',
                        'placeholder': 'token',
                        'defaultValue': (appData.data.settings.harvest && appData.data.settings.harvest.pass)
                    }
                ]
            }
        ];

        return (
            <div id="settings-container">
                <form onSubmit={this.submit}>
                    {forms.map((group, i) => {
                        return (
                            <div>
                                <h5>{group.group}</h5>
                                {group.forms.map((options, j) => {
                                    return (
                                        <div key={j} className="form-group">
                                            <label htmlFor={options.id}>{options.label}</label>
                                            <input
                                                type={options.type}
                                                className="form-control form-control-sm"
                                                id={options.id}
                                                placeholder={options.placeholder}
                                                defaultValue={options.defaultValue}
                                                aria-describedby={options.id + 'Help'}
                                            />
                                            {options.description && this.description(options)}
                                        </div>
                                    );
                                })}
                                <br />
                            </div>
                        );
                    })}

                    <button id="save" className="btn btn-primary btn-block" onClick={this.save}>Save</button>
                </form>
            </div>
        );
    }
}
