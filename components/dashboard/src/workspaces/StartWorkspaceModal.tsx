/**
 * Copyright (c) 2021 Gitpod GmbH. All rights reserved.
 * Licensed under the GNU Affero General Public License (AGPL).
 * See License.AGPL.txt in the project root for license information.
 */

import { StartOptions } from "@gitpod/gitpod-protocol/lib/util/gitpod-host-url";
import { useCallback, useContext, useMemo, useState } from "react";
import Modal from "../components/Modal";
import RepositoryFinder from "../components/RepositoryFinder";
import SelectIDEComponent from "../components/SelectIDEComponent";
import SelectWorkspaceClassComponent from "../components/SelectWorkspaceClassComponent";
import { UserContext } from "../user-context";

export interface StartWorkspaceModalProps {
    uselatestIde?: boolean;
    ide?: string;
    workspaceClass?: string;
    contextUrl?: string;
    onClose?: () => void;
}

export function StartWorkspaceModal(props: StartWorkspaceModalProps) {
    const { user } = useContext(UserContext);
    const [useLatestIde, setUseLatestIde] = useState<boolean | undefined>(
        props.uselatestIde || !!user?.additionalData?.ideSettings?.useLatestVersion,
    );
    const [selectedIde, setSelectedIde] = useState(props.ide || user?.additionalData?.ideSettings?.defaultIde);
    const [selectedWsClass, setSelectedWsClass] = useState<string | undefined>(props.workspaceClass);
    const [repo, setRepo] = useState<string | undefined>(props.contextUrl);
    const onSelectEditorChange = useCallback(
        (ide: string, useLatest: boolean) => {
            setSelectedIde(ide);
            setUseLatestIde(useLatest);
        },
        [setSelectedIde, setUseLatestIde],
    );

    const startWorkspace = useCallback(() => {
        if (!repo) {
            return false;
        }
        const url = new URL(window.location.href);
        url.pathname = "";
        const searchParams = new URLSearchParams();
        if (selectedWsClass) {
            searchParams.set(StartOptions.WORKSPACE_CLASS, selectedWsClass);
        }
        if (selectedIde) {
            searchParams.set(StartOptions.EDITOR, selectedIde);
            searchParams.set(StartOptions.USE_LATEST_EDITOR, useLatestIde ? "true" : "false");
        }
        url.search = searchParams.toString();
        url.hash = "#" + repo;
        window.location.href = url.toString();
        return true;
    }, [repo, selectedIde, selectedWsClass, useLatestIde]);

    const buttons = useMemo(() => {
        const result = [
            <button key="cancel" className="secondary" onClick={props.onClose}>
                Cancel
            </button>,
            <button key="start" className="" onClick={startWorkspace} disabled={!repo || repo.length === 0}>
                New Workspace
            </button>,
        ];
        if (!props.onClose) {
            return result.slice(1, 2);
        }
        return result;
    }, [props.onClose, repo, startWorkspace]);

    return (
        <Modal
            onClose={props.onClose || (() => {})}
            closeable={!!props.onClose}
            onEnter={startWorkspace}
            visible={true}
            title="Open in Gitpod"
            buttons={buttons}
        >
            <div className="-mx-6 px-6">
                <div className="text-xs text-gray-500">Start a new workspace with the following options.</div>
                <div className="pt-3">
                    <RepositoryFinder setSelection={props.contextUrl ? undefined : setRepo} initialValue={repo} />
                </div>
                <div className="pt-3">
                    <SelectIDEComponent
                        onSelectionChange={onSelectEditorChange}
                        selectedIdeOption={selectedIde}
                        useLatest={useLatestIde}
                    />
                </div>
                <div className="pt-3">
                    <SelectWorkspaceClassComponent
                        onSelectionChange={setSelectedWsClass}
                        selectedWorkspaceClass={selectedWsClass}
                    />
                </div>
            </div>
        </Modal>
    );
}
